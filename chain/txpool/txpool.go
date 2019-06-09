package txpool

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	ACore "github.com/ayachain/go-aya/consensus/core"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ayachain/go-aya/vdb"
	AAssets "github.com/ayachain/go-aya/vdb/assets"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	"github.com/ipfs/go-unixfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/whyrusleeping/go-logging"
	"sync"
	"time"
)

const AtxPoolVersion = "AyaTxPool 0.0.1"

var (
	ErrPackageThreadExpected 		= errors.New("an unrecoverable error occurred inside the thread")
	ErrRawDBEndoedZeroLen 			= errors.New("ready to sent content is empty")
	ErrMessageVerifyExpected 		= errors.New("message verify failed")
	ErrStorageRawDataDecodeExpected = errors.New("decode raw data expected")
	ErrStorageLowAPIExpected 		= errors.New("atx pool storage low api operation expected")
)

var(
	TxHashIteratorStart 	= []byte{0}
	TxHashIteratorLimit		= []byte{0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF}
	TxReceiptPrefix			= []byte{0xAE,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00}
	TxCount					= TxHashIteratorLimit
)


var log = logging.MustGetLogger("ATxPool")

/// Because nodes play different roles in the network, we divide the nodes into three categories:
/// super nodes, master nodes and ordinary nodes, and describe the operation of the transaction
/// pool which they are responsible for respectively.
///
/// According to the latest piece of information on the current chain, the trading pool will decide
/// its own mode of work and set up the switch to complete the mode of work.
type AtxPoolWorkMode uint8

const (
	/// The operations of common nodes for transaction pool include, but are not limited to, sending
	/// transactions, validating transactions, and saving some data.
	AtxPoolWorkModeNormal AtxPoolWorkMode = 0


	/// The main node's operations on the transaction pool include, but are not limited to, sending
	/// transactions, validating transactions, saving data, and computing ALVM execution results.
	AtxPoolWorkModeMaster AtxPoolWorkMode = 1


	/// The operation of the super node for the transaction pool includes, but is not limited to,
	/// sending transactions, validating transactions, saving data, blocking, packaging, computing
	/// ALVM execution results, etc.
	AtxPoolWorkModeSuper AtxPoolWorkMode = 2

	/// owner
	AtxPoolWorkModeOblivioned AtxPoolWorkMode = 3
)

type AtxThreadsName string

const (
	AtxThreadsAll					AtxThreadsName = "All"
	AtxThreadsNameTxPackage 		AtxThreadsName = "MiningBlockPackageThread"
	AtxThreadsNameBlockPacage		AtxThreadsName = "BlockConfirmPackageThread"
	AtxThreadsNameReceiptListen 	AtxThreadsName = "MiningBlockReceiptListenThread"
	AtxThreadsNameTransactionCommit AtxThreadsName = "TransactionCommitThread"
	AtxThreadsNameMining			AtxThreadsName = "MiningThread"
	AtxThreadsNameChannelListen		AtxThreadsName = "ChannelListenThread"
)

type ATxPool struct {

	cvfs vdb.CVFS
	storage *leveldb.DB
	ownerAccount EAccount.Account
	ownerAsset *AAssets.Assets

	adbpath string
	chainId string
	channelTopics string

	miningBlock *AMsgMBlock.MBlock
	miningBlockLocker sync.Mutex

	/// Top100 changes only when the block is updated. In order to reduce the reading and writing frequency
	/// of the database, we cache this information here and update it when the block is actually out.
	topAssets []*AAssets.SortAssets
	workmode AtxPoolWorkMode
	ind *core.IpfsNode

	poweron bool
	workctx context.Context
	workcancel context.CancelFunc

	threadCancels map[AtxThreadsName]context.CancelFunc
	threadChans map[AtxThreadsName] chan *AKeyStore.ASignedRawMsg

	notary ACore.Notary
	packLocker sync.Mutex
}


func NewTxPool( ctx context.Context, ind *core.IpfsNode, chainId string, cvfs vdb.CVFS, acc EAccount.Account) *ATxPool {

	adbpath := "/atxpool/" + chainId
	var nd *merkledag.ProtoNode
	dsk := datastore.NewKey(adbpath)
	val, err := ind.Repo.Datastore().Get(dsk)

	// create channel topices string
	topic := crypto.Keccak256Hash( []byte( AtxPoolVersion + chainId ) )

	switch {
	case err == datastore.ErrNotFound || val == nil:
		nd = unixfs.EmptyDirNode()

	case err == nil:

		c, err := cid.Cast(val)
		if err != nil {
			nd = unixfs.EmptyDirNode()
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second  * 5 )
		defer cancel()

		rnd, err := ind.DAG.Get(ctx, c)
		if err != nil {
			nd = unixfs.EmptyDirNode()
		}

		pbnd, ok := rnd.(*merkledag.ProtoNode)
		if !ok {
			nd = unixfs.EmptyDirNode()
		}

		nd = pbnd

	default:
		nd = unixfs.EmptyDirNode()
	}

	tops, err := cvfs.Assetses().GetLockedTop100()
	if err != nil {
		tops = []*AAssets.SortAssets{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second  * 5 )
	defer cancel()

	root, err := mfs.NewRoot(
		ctx,
		ind.DAG,
		nd,
		func(ctx context.Context, fcid cid.Cid) error {
			return ind.Repo.Datastore().Put(dsk, fcid.Bytes())
		},
	)

	dbnode, err := mfs.Lookup(root, "/" + chainId)
	if err != nil {

		if err := mfs.Mkdir(root, "/" + chainId, mfs.MkdirOpts{Flush:false, Mkparents:true}); err != nil {
			panic(err)
		}

	} else {

		dir, ok := dbnode.(*mfs.Directory)
		if !ok {
			goto configNewDir
		}

		db, err := leveldb.Open(ADB.NewMFSStorage(dir), nil)
		if err != nil {
			goto configNewDir
		}

		return &ATxPool{
			storage:db,
			cvfs:cvfs,
			topAssets:tops,
			workmode:AtxPoolWorkModeNormal,
			ind:ind,
			channelTopics:topic.String(),
			threadCancels:make(map[AtxThreadsName]context.CancelFunc),
			threadChans:make(map[AtxThreadsName] chan *AKeyStore.ASignedRawMsg),
		}
	}

configNewDir :

	newnode, err := mfs.Lookup(root, "/" + chainId)
	newdir, ok := newnode.(*mfs.Directory)
	if !ok {
		panic(mfs.ErrDirExists)
	}

	db, err := leveldb.Open(ADB.NewMFSStorage(newdir), nil)
	if err != nil {
		panic(mfs.ErrDirExists)
	}

	return &ATxPool{
		storage:db,
		cvfs:cvfs,
		topAssets:tops,
		workmode:AtxPoolWorkModeNormal,
		ind:ind,
		channelTopics:topic.String(),
		threadCancels:make(map[AtxThreadsName]context.CancelFunc),
		threadChans:make(map[AtxThreadsName] chan *AKeyStore.ASignedRawMsg),
	}

}

func (pool *ATxPool) PowerOn() error {

	if pool.poweron {
		return fmt.Errorf("%v ATxPool is already power on", pool.chainId)
	}

	pool.judgingMode()

	pool.workctx, pool.workcancel = context.WithCancel(context.Background())

	switch pool.workmode {

	case AtxPoolWorkModeOblivioned:

		pool.runthreads(
			AtxThreadsNameTxPackage,
			AtxThreadsNameBlockPacage,
			AtxThreadsNameReceiptListen,
			AtxThreadsNameTransactionCommit,
			AtxThreadsNameMining,
			AtxThreadsNameChannelListen,
		)

	case AtxPoolWorkModeSuper:

		pool.runthreads(
			AtxThreadsNameChannelListen,
			AtxThreadsNameTxPackage,
			AtxThreadsNameBlockPacage,
			AtxThreadsNameReceiptListen,
			AtxThreadsNameTransactionCommit,
			)


	case AtxPoolWorkModeMaster:

		pool.runthreads(
			AtxThreadsNameChannelListen,
			AtxThreadsNameReceiptListen,
			AtxThreadsNameTransactionCommit,
			AtxThreadsNameMining,
		)

	case AtxPoolWorkModeNormal:

		pool.runthreads(
			AtxThreadsNameChannelListen,
			AtxThreadsNameTransactionCommit,
		)

	}

	pool.poweron = true

	return nil
}

func (pool *ATxPool) runthreads( names ... AtxThreadsName ) {

	for _, n := range names {

		cancel, exist := pool.threadCancels[n]
		if exist && cancel != nil {
			cancel()
		}

		switch n {

		case AtxThreadsNameChannelListen:

			workCtx, cancel := context.WithCancel(pool.workctx)

			pool.threadCancels[n] = cancel
			pool.threadChans[n] = make(chan *AKeyStore.ASignedRawMsg)

			go pool.channelListening(workCtx)


		case AtxThreadsNameTxPackage:

			workCtx, cancel := context.WithCancel(pool.workctx)

			pool.threadCancels[n] = cancel
			pool.threadChans[n] = make(chan *AKeyStore.ASignedRawMsg)

			go pool.txPackageThread(workCtx)


		case AtxThreadsNameReceiptListen:

			workCtx, cancel := context.WithCancel(pool.workctx)

			pool.threadCancels[n] = cancel
			pool.threadChans[n] = make(chan *AKeyStore.ASignedRawMsg)

			go pool.receiptListen(workCtx)


		case AtxThreadsNameBlockPacage:

			workCtx, cancel := context.WithCancel(pool.workctx)

			pool.threadCancels[n] = cancel
			pool.threadChans[n] = make(chan *AKeyStore.ASignedRawMsg)

			go pool.blockPackageThread(workCtx)


		case AtxThreadsNameMining:

			workCtx, cancel := context.WithCancel(pool.workctx)

			pool.threadCancels[n] = cancel
			pool.threadChans[n] = make(chan *AKeyStore.ASignedRawMsg)

			go pool.miningThread(workCtx)

		}



	}

}

func (pool *ATxPool) PowerOff(err error) {
	log.Error(err)
	pool.workcancel()
	pool.poweron = false
}

func (pool *ATxPool) UpdateBestBlock( bcid cid.Cid ) error {

	pool.packLocker.Lock()
	defer pool.packLocker.Unlock()

	if err := pool.cvfs.SeekToBlock( bcid ); err != nil {
		return err
	}

	ast, err := pool.cvfs.Assetses().AssetsOf(pool.ownerAccount.Address.Bytes())
	if err != nil {
		return err
	}

	tops, err := pool.cvfs.Assetses().GetLockedTop100()
	if err != nil {
		tops = []*AAssets.SortAssets{}
	}

	pool.topAssets = tops
	pool.ownerAsset = ast

	return nil
}


func (pool *ATxPool) sign( coder AvdbComm.AMessageEncode ) (*AKeyStore.ASignedRawMsg, error) {

	cbs := coder.RawMessageEncode()

	if len(cbs) <= 0 {
		return nil, ErrRawDBEndoedZeroLen
	}

	return AKeyStore.CreateMsg(cbs, pool.ownerAccount)
}

/// Send a transaction by signing with the identity of the current node
func (pool *ATxPool) DoBroadcast( coder AvdbComm.AMessageEncode ) error {

	cbs := coder.RawMessageEncode()

	if len(cbs) <= 0 {
		return ErrRawDBEndoedZeroLen
	}

	if signmsg, err := AKeyStore.CreateMsg(cbs, pool.ownerAccount); err != nil {
		return err
	} else {

		rawsmsg, err := signmsg.Bytes()
		if err != nil {
			return err
		}

		return pool.ind.PubSub.Publish( pool.channelTopics, rawsmsg )
	}

}

/// We will record the miningblock receipts received by history. If the number of coins
/// used to prove the receipts is more than N times the number of coins held by the node
/// itself, it is admitted that this is the same for both the primary node and the ordinary
/// node.
func (pool *ATxPool) AddConfrimReceipt( mbhash EComm.Hash, retcid cid.Cid, from *EComm.Address) {

	/// If it is not possible to obtain the proof of "VoteRight" from the source, it means
	/// that the result has no reference value and is discarded directly.
	fromVoting, err := pool.cvfs.Assetses().VotingCountOf(from.Bytes())
	if err != nil {
		return
	}

	receiptKey := []byte(TxReceiptPrefix)
	copy(receiptKey[1:], mbhash.Bytes())

	exist, err := pool.storage.Has(receiptKey, nil)
	if err != nil {
		pool.Close()
		pool.KillByErr(ErrStorageLowAPIExpected)
	}

	rcidstr := retcid.String()
	receiptMap := make(map[string]uint64)
	if exist {

		value, err := pool.storage.Get(receiptKey, nil)
		if err != nil {
			pool.KillByErr(ErrStorageLowAPIExpected)
		}

		if json.Unmarshal(value, receiptMap) != nil {
			pool.KillByErr(ErrStorageRawDataDecodeExpected)
		}

		ocount, vexist := receiptMap[rcidstr]
		if vexist {
			receiptMap[rcidstr] = ocount + fromVoting
		} else {
			receiptMap[rcidstr] = fromVoting
		}

	} else {
		receiptMap[rcidstr] = fromVoting
	}

	defer func() {

		if receiptMap[rcidstr] > pool.ownerAsset.Vote * 3 {

			batch := &leveldb.Batch{}
			batch.Delete(receiptKey)
			batch.Delete(mbhash.Bytes())

			if err := pool.storage.Write(batch, nil); err != nil {
				pool.PowerOff(err)
			}

		}

	}()

	rmapbs, err := json.Marshal(receiptMap)
	if err != nil {
		pool.KillByErr(ErrStorageLowAPIExpected)
	}

	if err := pool.storage.Put( receiptKey, rmapbs, nil ); err != nil {
		pool.KillByErr(ErrStorageLowAPIExpected)
	}

}

func (pool *ATxPool) AddRawTransaction( tx *AKeyStore.ASignedRawMsg ) error {

	if !tx.Verify() {
		return ErrMessageVerifyExpected
	}

	txraw, err := tx.Bytes()
	if err != nil {
		return err
	}

	key := crypto.Keccak256(txraw)
	value := txraw

	if err := pool.storage.Put(TxCount, AvdbComm.BigEndianBytes(pool.Size() + 1), nil); err != nil {
		pool.PowerOff(err)
	}

	if err := pool.storage.Put(key, value, nil); err != nil {
		pool.PowerOff(err)
	}

	return nil
}

/// When an underlying invocation error occurs, the error cannot be handled or recovered,
/// and the program can only be killed before more errors are caused.
func (pool *ATxPool) KillByErr( err error ) {
	pool.Close()
	panic(err)
}


func (pool *ATxPool) Size() uint64 {

	if exist, err := pool.storage.Has(TxCount, nil); err != nil || !exist {
		return 0
	}

	indexbs, err := pool.storage.Get(TxCount, nil)

	if err != nil {
		panic(ErrStorageLowAPIExpected)
	}

	return binary.BigEndian.Uint64(indexbs)
}

func (pool *ATxPool) IsEmpty() bool {

	return pool.Size() == 0

}

func (pool *ATxPool) Close() {

	err := pool.storage.Close()

	if err != nil {
		log.Warning(err)
	}

}








/// private method
/// Judging working mode
func (pool *ATxPool) judgingMode() {

	pool.workmode = AtxPoolWorkModeOblivioned
	return

	n := 99999999

	for i := 0; i < len(pool.topAssets); i++ {
		if pool.topAssets[i] != nil && pool.topAssets[i].Addredd == pool.ownerAccount.Address {
			n = i
			break
		}
	}

	if n <= 21 {
		pool.workmode = AtxPoolWorkModeSuper
		return
	}

	oasset, err := pool.cvfs.Assetses().AssetsOf( pool.ownerAccount.Address.Bytes() )
	if err != nil {
		pool.workmode = AtxPoolWorkModeNormal
		return
	}

	if oasset.Locked > 100000 {
		pool.workmode = AtxPoolWorkModeMaster
		return
	}

	pool.workmode = AtxPoolWorkModeNormal
	return
}

func (pool *ATxPool) reduction() {

	txit := pool.storage.NewIterator(&util.Range{Start:TxHashIteratorStart, Limit:TxHashIteratorLimit},nil)
	defer txit.Release()

	batch := &leveldb.Batch{}

	for txit.Next() {

		// Remove the persistent backup if the transaction is ready for processing
		if pool.cvfs.Receipts().HasTransactionReceipt( EComm.BytesToHash(txit.Key()) ) {

			batch.Delete(txit.Key())

			receiptKey := []byte(TxReceiptPrefix)
			copy(receiptKey[1:], txit.Key())
			if ret, err := pool.storage.Has(receiptKey, nil); ret && err == nil {
				batch.Delete(receiptKey)
			}

		}

	}

	if batch.Len() > 0 {

		dbtransaction, err := pool.storage.OpenTransaction()
		if err != nil {
			pool.KillByErr(err)
		}

		if err := dbtransaction.Write(batch, nil); err != nil {
			pool.KillByErr(err)
		}

		if err := dbtransaction.Commit(); err != nil {
			pool.KillByErr(err)
		}

	}

}
