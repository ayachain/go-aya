package txpool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ayachain/go-aya/chain/txpool/txlist"
	ACore "github.com/ayachain/go-aya/consensus/core"
	"github.com/ayachain/go-aya/logs"
	"github.com/ayachain/go-aya/vdb"
	AAssets "github.com/ayachain/go-aya/vdb/assets"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	AElectoral "github.com/ayachain/go-aya/vdb/electoral"
	AMBlock "github.com/ayachain/go-aya/vdb/mblock"
	"github.com/ayachain/go-aya/vdb/node"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-ipfs/core"
	"github.com/whyrusleeping/go-logging"
	"sync"
	"time"
)

const AtxPoolVersion = "AyaTxPool 0.0.1"

var (
	ErrRawDBEndoedZeroLen 			= errors.New("ready to sent content is empty")
	ErrMessageVerifyExpected 		= errors.New("message verify failed")
	ErrTxVerifyExpected				= errors.New("tx tid verify expected")
	ErrTxVerifyInsufficientFunds	= errors.New("insufficient funds ( value + cost )")
)

var log = logging.MustGetLogger(logs.AModules_TxPool)

/// Because nodes play different roles in the network, we divide the nodes into three categories:
/// super nodes, master nodes and ordinary nodes, and describe the operation of the transaction
/// pool which they are responsible for respectively.
///
/// According to the latest piece of information on the current chain, the trading pool will decide
/// its own mode of work and set up the switch to complete the mode of work.
type AtxPoolWorkMode uint8

type ATxPoolThreadsName string

const (

	PackageTxsLimit 									= 2048

	AtxPoolWorkModeNormal 			AtxPoolWorkMode 	= 0

	AtxPoolWorkModeSuper 			AtxPoolWorkMode 	= 1

	ATxPoolThreadTxListen			ATxPoolThreadsName 	= "thread.tx.listen"
	AtxPoolThreadTxListenBuff 					   		= 128

	ATxPoolThreadTxPackage 			ATxPoolThreadsName 	= "thread.tx.package"
	ATxPoolThreadTxPackageBuff					  		= 1

	ATxPoolThreadExecutor			ATxPoolThreadsName 	= "thread.block.executor"
	ATxPoolThreadExecutorBuff					  		= 8

	ATxPoolThreadReceiptListen 		ATxPoolThreadsName 	= "thread.receipt.listen"
	ATxPoolThreadReceiptListenBuff						= 256

	ATxPoolThreadMining				ATxPoolThreadsName 	= "thread.block.mining"
	ATxPoolThreadMiningBuff								= 32

	ATxPoolThreadChainInfo			ATxPoolThreadsName 	= "thread.chaininfo.listen"
	ATxPoolThreadChainInfoBuff							= 16

	ATxPoolThreadElectoral			ATxPoolThreadsName	= "thread.packer.electoral"
	ATxPoolThreadElectoralBuff							= 21

	ATxPoolThreadElectoralTimeout	ATxPoolThreadsName 	= "thread.packer.timeout"
)

type ATxPool struct {

	cvfs vdb.CVFS

	mining map[EComm.Address]*txlist.TxList
	queue map[EComm.Address]*txlist.TxList

	ownerAccount EAccount.Account
	ownerAsset *AAssets.Assets
	genBlock *ABlock.GenBlock
	miningBlock *AMBlock.MBlock

	workmode AtxPoolWorkMode
	ind *core.IpfsNode

	channelTopics map[ATxPoolThreadsName] string

	threadChans sync.Map
	notary ACore.Notary
	workingThreadWG sync.WaitGroup

	syncMutx sync.Mutex

	packerState AElectoral.ATxPackerState
	latestPackerStateChangeTime int64
	eleservices AElectoral.MemServices
}

func NewTxPool( ind *core.IpfsNode, gblk *ABlock.GenBlock, cvfs vdb.CVFS, miner ACore.Notary, acc EAccount.Account) *ATxPool {

	// create channel topices string
	topic := crypto.Keccak256Hash( []byte( AtxPoolVersion + gblk.ChainID ) )

	topicmap := map[ATxPoolThreadsName]string{
		ATxPoolThreadTxListen 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxListen))).String(),
		ATxPoolThreadTxPackage 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxPackage))).String(),
		ATxPoolThreadExecutor 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadExecutor))).String(),
		ATxPoolThreadReceiptListen 	: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadReceiptListen))).String(),
		ATxPoolThreadMining 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadMining))).String(),
		ATxPoolThreadChainInfo		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadChainInfo))).String(),
		ATxPoolThreadElectoral 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadElectoral))).String(),
	}

	oast, err := cvfs.Assetses().AssetsOf(acc.Address)
	if err != nil {
		oast = &AAssets.Assets{
			Version:AAssets.DRVer,
			Avail:0,
			Locked:0,
			Vote:0,
		}
	}

	return &ATxPool{
		mining:make(map[EComm.Address]*txlist.TxList),
		queue:make(map[EComm.Address]*txlist.TxList),
		cvfs:cvfs,
		workmode:AtxPoolWorkModeNormal,
		ind:ind,
		channelTopics:topicmap,
		ownerAccount:acc,
		ownerAsset:oast,
		notary:miner,
		genBlock:gblk,
		packerState:AElectoral.ATxPackStateLookup,
		latestPackerStateChangeTime:time.Now().Unix(),
		eleservices:AElectoral.CreateServices(cvfs, 10),
	}

}

func (pool *ATxPool) PowerOn( ctx context.Context ) error {

	pool.judgingMode()

	fmt.Println("ATxPool Working Started.")

	defer fmt.Println("ATxPool Working Stoped.")

	switch pool.workmode {

	case AtxPoolWorkModeSuper:

		pool.runThreads(
			ctx,
			ATxPoolThreadElectoral,
			ATxPoolThreadTxListen,
			ATxPoolThreadTxPackage,
			ATxPoolThreadMining,
			ATxPoolThreadReceiptListen,
			ATxPoolThreadExecutor,
			ATxPoolThreadChainInfo,
		)

	case AtxPoolWorkModeNormal:

		pool.runThreads(
			ctx,
			//ATxPoolThreadElectoral,
			//ATxPoolThreadTxListen,
			//ATxPoolThreadTxPackage,
			//ATxPoolThreadMining,
			//ATxPoolThreadReceiptListen,
			ATxPoolThreadExecutor,
			ATxPoolThreadChainInfo,
		)

	}

	select {

	case <- ctx.Done():

		pool.workingThreadWG.Wait()

		return ctx.Err()
	}
}

func (pool *ATxPool) UpdateBestBlock( cblock *ABlock.Block  ) error {

	ast, err := pool.cvfs.Assetses().AssetsOf( pool.ownerAccount.Address )

	if err != nil {
		return err
	}

	pool.ownerAsset = ast

	return nil
}

func (pool *ATxPool) DoPackMBlock() {

	cc, exist := pool.threadChans.Load(ATxPoolThreadTxPackage)

	if !exist {
		return
	}

	cc.(chan []byte) <- nil

	return
}

func (pool *ATxPool) ReadOnlyCVFS() vdb.CVFS {
	return pool.cvfs
}

func (pool *ATxPool) PublishTx( tx *ATx.Transaction ) error {

	if !tx.Verify() {
		return ErrMessageVerifyExpected
	}

	return pool.doBroadcast(tx, pool.channelTopics[ATxPoolThreadTxListen])
}

func (pool *ATxPool) ElectoralServices() AElectoral.MemServices {
	return pool.eleservices
}

func (pool *ATxPool) GetWorkMode() AtxPoolWorkMode {
	return pool.workmode
}

func (pool *ATxPool) GetState() *State {

	queueSum := 0
	for k := range pool.queue {
		queueSum += pool.queue[k].Len()
	}

	pendingSum := 0
	for k := range pool.mining {
		pendingSum += pool.mining[k].Len()
	}

	s := &State{
		Account	: pool.ownerAccount.Address.String(),
		Queue	: queueSum,
		Pending	: pendingSum,
		Version	: AtxPoolVersion,
	}

	if pool.workmode == AtxPoolWorkModeSuper {

		s.WorkMode = "Super"

	} else {

		s.WorkMode = "Normal"

	}

	return s
}

//func (pool *ATxPool) TxExist( hash EComm.Hash ) (exist bool, mname string) {
//
//
//
//}

func (pool *ATxPool) PushTransaction( tx *ATx.Transaction ) error {

	if !tx.Verify() {
		return ErrMessageVerifyExpected
	}

	/// verify tid
	txsum, err := pool.cvfs.Transactions().GetTxCount(tx.From)
	if err != nil || tx.Tid < txsum {
		return ErrTxVerifyExpected
	}

	if list, exist := pool.queue[tx.From]; !exist {

		pool.queue[tx.From] = txlist.NewTxList(tx)

		return nil

	} else {

		return list.AddTx(tx)
	}

}

func (pool *ATxPool) MoveTxsToMining( txs []*ATx.Transaction ) error {

	for _, stx := range txs {

		if tlist, exist := pool.queue[stx.From]; exist {

			tlist.RemoveFromTid( stx.Tid )

			if plist, pexist := pool.mining[stx.From]; pexist {

				_ = plist.AddTx(stx)

			} else {

				pool.mining[stx.From] = txlist.NewTxList(stx)

			}

		}

	}

	return nil

}

func (pool *ATxPool) ConfirmTxs( txs []*ATx.Transaction ) error {

	for _, stx := range txs {

		removed := false

		if tlist, exist := pool.mining[stx.From]; exist {

			removed = tlist.RemoveFromTid( stx.Tid )

		}

		if !removed {

			if tlist, exist := pool.queue[stx.From]; exist {

				_ = tlist.RemoveFromTid( stx.Tid )

			}

		}

	}

	return nil

}

func (pool *ATxPool) CreateMiningBlock() *AMBlock.MBlock {

	if pool.packerState != AElectoral.ATxPackStateMaster {
		return nil
	}

	var packtxs []*ATx.Transaction

	for addr, txl := range pool.queue {

		if txcount, err := pool.cvfs.Transactions().GetTxCount(addr); err != nil {

			continue

		} else {

			if ftx := txl.FrontTx(); ftx != nil && ftx.Tid == txcount {

				//can packing to this mining block
				if txs := txl.GetLinearTxsFromFront(); txs != nil {

					packtxs = append(packtxs, txs...)

					if len(packtxs) > PackageTxsLimit {
						break
					}

				}

			}

		}

	}

	if len(packtxs) < 0 {
		return nil
	}

	txsblockcontent, err := json.Marshal(packtxs)
	if err != nil {
		log.Error(err)
		return nil
	}

	iblk := blocks.NewBlock(txsblockcontent)
	err = pool.ind.Blocks.AddBlock(iblk)
	if err != nil {
		log.Error(err)
		return nil
	}

	// Create block
	bindex, err := pool.cvfs.Indexes().GetLatest()
	if err != nil {
		log.Error(err)
		return nil
	}

	mblk := &AMBlock.MBlock{}
	mblk.ExtraData = ""
	mblk.Index = bindex.BlockIndex + 1
	mblk.ChainID = pool.genBlock.ChainID
	mblk.Parent = bindex.Hash.String()
	mblk.Timestamp = uint64(time.Now().Unix())
	mblk.Packager = pool.ownerAccount.Address.String()
	mblk.Txc = uint16(len(packtxs))
	mblk.Txs = iblk.Cid().String()

	return mblk
}

/// Private method
/// Judging working mode
func (pool *ATxPool) judgingMode() {

	nd, err := pool.cvfs.Nodes().GetNodeByPeerId( pool.ind.Identity.Pretty() )

	if err != nil {

		pool.workmode = AtxPoolWorkModeNormal
		log.Error("TxPool WorkMode [Normal]")

	} else if nd.Type == node.NodeTypeSuper {

		pool.workmode = AtxPoolWorkModeSuper
		log.Info("TxPool WorkMode [Super]")

	} else {

		pool.workmode = AtxPoolWorkModeNormal
		log.Error("TxPool WorkMode [Normal]")

	}

	return
}

func (pool *ATxPool) runThreads( ctx context.Context, names ... ATxPoolThreadsName ) {

	for _, n := range names {

		sctx := context.WithValue( ctx, "Pool", pool )

		switch n {

		case ATxPoolThreadTxListen:
			go txListenThread(sctx)


		case ATxPoolThreadTxPackage:
			go txPackageThread(sctx)


		case ATxPoolThreadMining:
			go miningThread(sctx)


		case ATxPoolThreadReceiptListen:
			go receiptListen(sctx)


		case ATxPoolThreadExecutor:
			go blockExecutorThread(sctx)


		case ATxPoolThreadChainInfo:
			go syncListener(sctx)


		case ATxPoolThreadElectoral:
			go electoralThread(sctx)

		}

	}

}

func (pool *ATxPool) doBroadcast( coder AvdbComm.AMessageEncode, topic string) error {

	cbs := coder.RawMessageEncode()

	if len(cbs) <= 0 {
		return ErrRawDBEndoedZeroLen
	}

	return pool.ind.PubSub.Publish( topic, cbs )

}

func (pool *ATxPool) changePackerState( s AElectoral.ATxPackerState ) {

	pool.latestPackerStateChangeTime = time.Now().Unix()
	pool.packerState = s

}