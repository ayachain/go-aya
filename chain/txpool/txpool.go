package txpool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ayachain/go-aya/chain/txpool/txlist"
	"github.com/ayachain/go-aya/vdb"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	AElectoral "github.com/ayachain/go-aya/vdb/electoral"
	"github.com/ayachain/go-aya/vdb/indexes"
	AMBlock "github.com/ayachain/go-aya/vdb/mblock"
	"github.com/ayachain/go-aya/vdb/node"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/pin"
	"github.com/prometheus/common/log"
	"sync"
	"time"
)

const AtxPoolVersion = "AyaTxPool 0.0.1"

var (
	ErrRawDBEndoedZeroLen 			= errors.New("ready to sent content is empty")
	ErrMessageVerifyExpected 		= errors.New("message verify failed")
	ErrTxVerifyExpected				= errors.New("tx tid verify expected")
	ErrTxVerifyInsufficientFunds	= errors.New("insufficient funds ( value + cost )")

	ErrCannotCreateMiningBlock		= errors.New("can not create new mining block")
)

type AtxPoolWorkMode uint8
type ATxPoolThreadsName string
type ATxPoolQueueName string

const (

	ATxPoolQueueNameUnknown			ATxPoolQueueName	= "Unknown"
	ATxPoolQueueNameMining			ATxPoolQueueName	= "MiningPool"
	ATxPoolQueueNameQueue			ATxPoolQueueName  	= "QueuePool"

	PackageTxsLimit 									= 2048

	AtxPoolWorkModeNormal 			AtxPoolWorkMode 	= 0
	AtxPoolWorkModeSuper 			AtxPoolWorkMode 	= 1

	ATxPoolThreadTxListen			ATxPoolThreadsName 	= "ATxPoolThreadTxListen"
	ATxPoolThreadTxPackage 			ATxPoolThreadsName 	= "ATxPoolThreadTxPackage"
	ATxPoolThreadElectoral			ATxPoolThreadsName	= "ATxPoolThreadElectoral"

	ATxPoolThreadElectoralTimeout	ATxPoolThreadsName 	= "ATxPoolThreadElectoralTimeout"
)

type ATxPool struct {

	chainID string
	ind *core.IpfsNode
	idxServices indexes.IndexesServices
	cvfs vdb.CVFS

	mnmu sync.Mutex
	qemu sync.Mutex

	mining map[EComm.Address]*txlist.TxList
	queue map[EComm.Address]*txlist.TxList

	channelTopics map[ATxPoolThreadsName] string

	workmode AtxPoolWorkMode
	ownerAccount EAccount.Account
	miningBlock *AMBlock.MBlock

	packerState AElectoral.ATxPackerState
	latestPackerStateChangeTime int64
	eleservices AElectoral.MemServices

	miningBlockBroadcastFun func(mblock *AMBlock.MBlock) error
}

func NewTxPool( ind *core.IpfsNode, chainID string, cvfs vdb.CVFS, acc EAccount.Account, mbfun func(mblock *AMBlock.MBlock) error ) *ATxPool {

	// create channel topices string
	topic := crypto.Keccak256Hash( []byte( AtxPoolVersion + chainID ) )
	topicmap := map[ATxPoolThreadsName]string{
		ATxPoolThreadTxListen 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxListen))).String(),
		ATxPoolThreadTxPackage 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxPackage))).String(),
		ATxPoolThreadElectoral 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadElectoral))).String(),
	}

	return &ATxPool{
		mining:make(map[EComm.Address]*txlist.TxList),
		queue:make(map[EComm.Address]*txlist.TxList),
		cvfs:cvfs,
		workmode:AtxPoolWorkModeNormal,
		ind:ind,
		channelTopics:topicmap,
		ownerAccount:acc,
		chainID:chainID,
		packerState:AElectoral.ATxPackStateLookup,
		latestPackerStateChangeTime:time.Now().Unix(),
		eleservices:AElectoral.CreateServices(cvfs, 10),
		miningBlockBroadcastFun:mbfun,
	}
}

func (pool *ATxPool) PowerOn( ctx context.Context ) {

	log.Info("ATxPool PowerOn")
	defer log.Info("ATxPool PowerOff")

	switch pool.judgingMode() {
	case AtxPoolWorkModeSuper:

		ctx1, cancel1 := context.WithCancel(ctx)
		ctx2, cancel2 := context.WithCancel(ctx)
		defer cancel1()
		defer cancel2()

		go pool.threadTransactionListener(ctx1)
		go pool.threadElectoralAndPacker(ctx2)

		select {
		case <- ctx.Done():
			return

		case <- ctx1.Done():
			return

		case <- ctx2.Done():
			return
		}

	default :

		ctx1, cancel1 := context.WithCancel(ctx)
		defer cancel1()

		go pool.threadTransactionListener(ctx1)

		select {
		case <- ctx.Done():
			return

		case <- ctx1.Done():
			return
		}
	}
}

func (pool *ATxPool) ConfirmBestBlock( cblock *ABlock.Block  ) error {

	// clear txpool
	dagReadCtx, dagReadCancel := context.WithCancel(context.TODO())

	confirmTxlist := cblock.ReadTxsFromDAG(dagReadCtx, pool.ind)

	dagReadCancel()

	if err := pool.confirmTxs( confirmTxlist ); err != nil {
		log.Error(err)
		return err
	}

	pool.changePackerState(AElectoral.ATxPackStateLookup)

	pool.miningBlock = nil

	return nil
}

func (pool *ATxPool) PublishTx( tx *ATx.Transaction ) error {

	if !tx.Verify() {
		return ErrMessageVerifyExpected
	}

	return pool.doBroadcast(tx, pool.channelTopics[ATxPoolThreadTxListen])
}

func (pool *ATxPool) DoPackMBlock() error {

	mblock := pool.createMiningBlock()
	if mblock == nil {
		return ErrCannotCreateMiningBlock
	}

	return pool.miningBlockBroadcastFun(mblock)
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

	miningSum := 0
	for k := range pool.mining {
		miningSum += pool.mining[k].Len()
	}

	s := &State{
		Account	: pool.ownerAccount.Address.String(),
		Queue	: queueSum,
		Mining	: miningSum,
		Version	: AtxPoolVersion,
	}

	if pool.workmode == AtxPoolWorkModeSuper {

		s.WorkMode = "Super"

	} else {

		s.WorkMode = "Normal"

	}

	return s
}

func (pool *ATxPool) TxExist( hash EComm.Hash ) (exist bool, mname ATxPoolQueueName) {

	pool.qemu.Lock()
	defer pool.qemu.Unlock()

	pool.mnmu.Lock()
	defer pool.mnmu.Unlock()

	for _, tlist := range pool.mining {

		if tlist.Exist(hash) {

			return true, ATxPoolQueueNameMining

		}

	}

	for _, tlist := range pool.queue {

		if tlist.Exist(hash) {

			return true, ATxPoolQueueNameQueue

		}


	}

	return false, ATxPoolQueueNameUnknown

}

func (pool *ATxPool) GetTx( hash EComm.Hash, mname ATxPoolQueueName ) *ATx.Transaction {

	switch mname {

	case ATxPoolQueueNameMining :

		pool.qemu.Lock()
		defer pool.qemu.Unlock()

		for _, tlist := range pool.mining {

			if tlist.Exist(hash) {

				return tlist.Get(hash)

			}

		}

		return nil


	case ATxPoolQueueNameQueue :

		pool.mnmu.Lock()
		defer pool.mnmu.Unlock()

		for _, tlist := range pool.mining {

			if tlist.Exist(hash) {

				return tlist.Get(hash)

			}

		}
	}

	return nil

}

func (pool *ATxPool) MoveTxsToMining( txs []*ATx.Transaction ) error {

	pool.qemu.Lock()
	defer pool.qemu.Unlock()

	pool.mnmu.Lock()
	defer pool.mnmu.Unlock()

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


/// Private methods
func (pool *ATxPool) createMiningBlock() *AMBlock.MBlock {

	pool.qemu.Lock()
	defer pool.qemu.Unlock()

	pool.mnmu.Lock()
	defer pool.mnmu.Unlock()

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
	pool.ind.Pinning.PinWithMode(iblk.Cid(), pin.Any)

	// Create block
	bindex, err := pool.cvfs.Indexes().GetLatest()
	if err != nil {
		log.Error(err)
		return nil
	}

	mblk := &AMBlock.MBlock{}
	mblk.ExtraData = ""
	mblk.Index = bindex.BlockIndex + 1
	mblk.ChainID = pool.chainID
	mblk.Parent = bindex.Hash.String()
	mblk.Timestamp = uint64(time.Now().Unix())
	mblk.Packager = pool.ownerAccount.Address.String()
	mblk.Txc = uint16(len(packtxs))
	mblk.Txs = iblk.Cid()

	return mblk
}

func (pool *ATxPool) judgingMode() AtxPoolWorkMode {

	nd, err := pool.cvfs.Nodes().GetNodeByPeerId( pool.ind.Identity.Pretty() )

	if err != nil {

		pool.workmode = AtxPoolWorkModeNormal

	} else if nd.Type == node.NodeTypeSuper {

		pool.workmode = AtxPoolWorkModeSuper

	} else {

		pool.workmode = AtxPoolWorkModeNormal
	}

	return pool.workmode
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

func (pool *ATxPool) storeTransaction( tx *ATx.Transaction ) error {

	pool.qemu.Lock()
	defer pool.qemu.Unlock()

	pool.mnmu.Lock()
	defer pool.mnmu.Unlock()

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

func (pool *ATxPool) confirmTxs( txs []*ATx.Transaction ) error {

	pool.qemu.Lock()
	defer pool.qemu.Unlock()

	pool.mnmu.Lock()
	defer pool.mnmu.Unlock()

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