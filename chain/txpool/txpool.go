package txpool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ACore "github.com/ayachain/go-aya/consensus/core"
	"github.com/ayachain/go-aya/logs"
	"github.com/ayachain/go-aya/vdb"
	AAssets "github.com/ayachain/go-aya/vdb/assets"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	AElectoral "github.com/ayachain/go-aya/vdb/electoral"
	AMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	"github.com/ayachain/go-aya/vdb/node"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/whyrusleeping/go-logging"
	"sync"
	"time"
)

const AtxPoolVersion = "AyaTxPool 0.0.1"

var (
	ErrRawDBEndoedZeroLen 			= errors.New("ready to sent content is empty")
	ErrMessageVerifyExpected 		= errors.New("message verify failed")
)

var(
	TxHashIteratorStart 	= []byte{0}
	TxHashIteratorLimit		= []byte{0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF}
	TxReceiptPrefix			= []byte{0xAE,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00}
	//TxCount					= TxHashIteratorLimit
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

	storage *leveldb.DB

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

	txlenmu sync.Mutex
	txlen uint
}

func NewTxPool( ind *core.IpfsNode, gblk *ABlock.GenBlock, cvfs vdb.CVFS, miner ACore.Notary, acc EAccount.Account) *ATxPool {

	// create channel topices string
	topic := crypto.Keccak256Hash( []byte( AtxPoolVersion + gblk.ChainID ) )

	topicmap := map[ATxPoolThreadsName]string{
		ATxPoolThreadTxListen : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxListen))).String(),
		ATxPoolThreadTxPackage : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxPackage))).String(),
		ATxPoolThreadExecutor : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadExecutor))).String(),
		ATxPoolThreadReceiptListen : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadReceiptListen))).String(),
		ATxPoolThreadMining : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadMining))).String(),
		ATxPoolThreadChainInfo: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadChainInfo))).String(),
		ATxPoolThreadElectoral : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadElectoral))).String(),
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

	memstore := storage.NewMemStorage()
	db, err := leveldb.Open(memstore, AvdbComm.OpenDBOpt)

	return &ATxPool{
		storage:db,
		cvfs:cvfs,
		workmode:AtxPoolWorkModeNormal,
		ind:ind,
		channelTopics:topicmap,
		ownerAccount:acc,
		ownerAsset:oast,
		notary:miner,
		genBlock:gblk,
		packerState:AElectoral.ATxPackStateUnknown,
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

		if err := pool.storage.Close(); err != nil {
			log.Error(err)
		}

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

	s := &State{
		Account: pool.ownerAccount.Address.String(),
		Len:     uint64(pool.txlen),
		MemorySize:0,
		WorkMode: "Unknown",
	}

	sizes, err := pool.storage.SizeOf( []util.Range{ {nil,nil} } )
	if err == nil {
		s.MemorySize = sizes.Sum()
	}

	if pool.workmode == AtxPoolWorkModeSuper {

		s.WorkMode = "Super"

	} else {

		s.WorkMode = "Normal"

	}

	return s
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

func (pool *ATxPool) addRawTransaction( tx *ATx.Transaction ) error {

	pool.txlenmu.Lock()
	defer pool.txlenmu.Unlock()

	if !tx.Verify() {
		return ErrMessageVerifyExpected
	}

	txraw := tx.Encode()
	if len(txraw) <= 0 {
		return ErrMessageVerifyExpected
	}

	key := crypto.Keccak256(txraw)
	if err := pool.storage.Put(key, txraw, nil); err != nil {
		return err
	}

	pool.txlen ++

	pool.DoPackMBlock()

	return nil
}

func (pool *ATxPool) removeExistedTxsFromMiningBlock( mblock *AMsgMBlock.MBlock ) error {

	pool.txlenmu.Lock()
	defer pool.txlenmu.Unlock()

	txsCid, err := cid.Decode(mblock.Txs)
	if err != nil {
		return err
	}

	iblock, err := pool.ind.Blocks.GetBlock(context.TODO(), txsCid)
	if err != nil {
		return err
	}

	txlist := make([]*ATx.Transaction, mblock.Txc)
	if err := json.Unmarshal(iblock.RawData(), &txlist); err != nil {
		return err
	}

	for _, tx := range txlist {

		txraw := tx.Encode()
		if len(txraw) <= 0 {
			continue
		}

		key := crypto.Keccak256(txraw)

		if exist, err := pool.storage.Has(key, nil); exist && err == nil {

			if err := pool.storage.Delete( key, nil ); err != nil {
				log.Warning(err)
				continue
			}

			pool.txlen --

		}

	}

	return nil
}

func (pool *ATxPool) changePackerState( s AElectoral.ATxPackerState ) {

	pool.latestPackerStateChangeTime = time.Now().Unix()
	pool.packerState = s

}