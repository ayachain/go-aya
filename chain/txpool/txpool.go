package txpool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ASD "github.com/ayachain/go-aya/chain/sdaemon/common"
	"github.com/ayachain/go-aya/chain/txpool/txlist"
	"github.com/ayachain/go-aya/vdb"
	AElectoral "github.com/ayachain/go-aya/vdb/electoral"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/indexes"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
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

const (

	PackageTxsLimit 									= 2048

	AtxPoolWorkModeNormal 			AtxPoolWorkMode 	= 0
	AtxPoolWorkModeSuper 			AtxPoolWorkMode 	= 1

	ATxPoolThreadTxListen			ATxPoolThreadsName 	= "ATxPoolThreadTxListen"
	ATxPoolThreadTxPackage 			ATxPoolThreadsName 	= "ATxPoolThreadTxPackage"
	ATxPoolThreadRepeater			ATxPoolThreadsName	= "ATxPoolThreadRepeater"

	ATxPoolThreadElectoralTimeout	ATxPoolThreadsName 	= "ATxPoolThreadElectoralTimeout"
)

type TxPool interface {

	PowerOn( ctx context.Context )

	PublishTx( tx *im.Transaction ) error

	ElectoralServices() AElectoral.MemServices

	GetWorkMode() AtxPoolWorkMode

	GetState() *State

	GetTx( hash EComm.Hash ) *im.Transaction
}

type aTxPool struct {

	chainID string
	ind *core.IpfsNode
	idxServices indexes.IndexesServices
	cvfs vdb.CVFS

	pmu sync.Mutex
	pending map[EComm.Address]*txlist.TxList
	channelTopics map[ATxPoolThreadsName] string

	workmode AtxPoolWorkMode
	ownerAccount EAccount.Account

	eleservices AElectoral.MemServices

	mblockChannel string
	asd ASD.StatDaemon

	lmblock *im.Block
}

func NewTxPool( ind *core.IpfsNode, chainID string, cvfs vdb.CVFS, acc EAccount.Account, mchannel string, asd ASD.StatDaemon ) TxPool {

	// create channel topices string
	topic := crypto.Keccak256Hash( []byte( AtxPoolVersion + chainID ) )
	topicmap := map[ATxPoolThreadsName]string{
		ATxPoolThreadTxListen 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxListen))).String(),
		ATxPoolThreadTxPackage 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxPackage))).String(),
		ATxPoolThreadRepeater 		: crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadRepeater))).String(),
	}

	return &aTxPool{
		pending:make(map[EComm.Address]*txlist.TxList),
		cvfs:cvfs,
		workmode:AtxPoolWorkModeNormal,
		ind:ind,
		channelTopics:topicmap,
		ownerAccount:acc,
		chainID:chainID,
		eleservices:AElectoral.CreateServices(cvfs, 10),
		mblockChannel:mchannel,
		asd:asd,
	}
}

func (pool *aTxPool) PowerOn( ctx context.Context ) {

	log.Info("TXP On")
	defer log.Info("TXP Off")

	wg := &sync.WaitGroup{}

	switch pool.judgingMode() {
	case AtxPoolWorkModeSuper:

		ctx1, cancel1 := context.WithCancel(ctx)
		ctx2, cancel2 := context.WithCancel(ctx)
		ctx3, cancel3 := context.WithCancel(ctx)

		go pool.threadTransactionListener(ctx1, wg)
		go pool.threadElectoralAndPacker(ctx2, wg)
		go pool.threadMiningBlockRepeater(ctx3, wg)

		select {
		case <- ctx.Done():
			break

		case <- ctx1.Done():
			break

		case <- ctx2.Done():
			break

		case <- ctx3.Done():
			break
		}

		cancel1()
		cancel2()
		cancel3()

		wg.Wait()

	default :

		ctx1, cancel1 := context.WithCancel(ctx)

		go pool.threadTransactionListener(ctx1, wg)

		select {
		case <- ctx.Done():
			break

		case <- ctx1.Done():
			break
		}

		cancel1()
		wg.Wait()
	}
}

func (pool *aTxPool) PublishTx( tx *im.Transaction ) error {

	if !tx.Verify() {
		return ErrMessageVerifyExpected
	}

	return pool.doBroadcast(tx, pool.channelTopics[ATxPoolThreadTxListen])
}

func (pool *aTxPool) ElectoralServices() AElectoral.MemServices {
	return pool.eleservices
}

func (pool *aTxPool) GetWorkMode() AtxPoolWorkMode {
	return pool.workmode
}

func (pool *aTxPool) GetState() *State {

	pendingSum := 0
	for k := range pool.pending {
		pendingSum += pool.pending[k].Len()
	}

	s := &State{
		Account	: pool.ownerAccount.Address.String(),
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

func (pool *aTxPool) GetTx( hash EComm.Hash ) *im.Transaction {

	pool.pmu.Lock()
	defer pool.pmu.Unlock()

	for _, tlist := range pool.pending {

		if tlist.Exist(hash) {
			return tlist.Get(hash)
		}

	}

	return nil

}

/// Private methods
func (pool *aTxPool) createMiningBlock() *im.Block {

	pool.pmu.Lock()
	defer pool.pmu.Unlock()

	var packtxs []*im.Transaction

	for addr, txl := range pool.pending {

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

	mblk := &im.Block{}
	mblk.ExtraData = nil
	mblk.Index = bindex.BlockIndex + 1
	mblk.ChainID = pool.chainID
	mblk.Parent = bindex.Hash.String()
	mblk.Timestamp = uint32(time.Now().Unix())
	mblk.Packager = pool.ownerAccount.Address.String()
	mblk.Txc = uint32(len(packtxs))
	mblk.Txs = iblk.Cid().Bytes()

	return mblk
}

func (pool *aTxPool) judgingMode() AtxPoolWorkMode {

	nd, err := pool.cvfs.Nodes().GetNodeByPeerId( pool.ind.Identity.Pretty() )

	if err != nil {

		pool.workmode = AtxPoolWorkModeNormal

	} else if nd.Type == im.NodeType_Super {

		pool.workmode = AtxPoolWorkModeSuper

	} else {

		pool.workmode = AtxPoolWorkModeNormal
	}

	return pool.workmode
}

func (pool *aTxPool) doBroadcast( pmsg proto.Message, topic string) error {

	cbs, err := proto.Marshal(pmsg)
	if err != nil {
		return err
	}

	if len(cbs) <= 0 {
		return ErrRawDBEndoedZeroLen
	}

	return pool.ind.PubSub.Publish( topic, cbs )
}

func (pool *aTxPool) storeTransaction( tx *im.Transaction ) error {

	pool.pmu.Lock()
	defer pool.pmu.Unlock()

	if !tx.Verify() {
		return ErrMessageVerifyExpected
	}

	/// verify tid
	txsum, err := pool.cvfs.Transactions().GetTxCount( EComm.BytesToAddress(tx.From) )
	if err != nil || tx.Tid < txsum {
		return ErrTxVerifyExpected
	}

	if list, exist := pool.pending[ EComm.BytesToAddress(tx.From) ]; !exist {

		pool.pending[ EComm.BytesToAddress(tx.From) ] = txlist.NewTxList(tx)

		return nil

	} else {

		return list.AddTx(tx)
	}

}

func (pool *aTxPool) confirmTxs( txs []*im.Transaction ) error {

	pool.pmu.Lock()
	defer pool.pmu.Unlock()

	for _, stx := range txs {

		if tlist, exist := pool.pending[ EComm.BytesToAddress(stx.From) ]; exist {
			_ = tlist.RemoveFromTid( stx.Tid )
		}

	}

	return nil
}

func (pool *aTxPool) doPackMBlock() (*im.Block, error) {

	mblock := pool.createMiningBlock()
	if mblock == nil {
		return nil, ErrCannotCreateMiningBlock
	}

	cbs,err := proto.Marshal(mblock)
	if err != nil {
		return nil, err
	}

	if len(cbs) <= 0 {
		return nil, ErrRawDBEndoedZeroLen
	}

	err = pool.ind.PubSub.Publish( pool.channelTopics[ATxPoolThreadRepeater], cbs )

	return mblock, err
}