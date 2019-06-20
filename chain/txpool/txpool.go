package txpool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	ACore "github.com/ayachain/go-aya/consensus/core"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ayachain/go-aya/vdb"
	AAssets "github.com/ayachain/go-aya/vdb/assets"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	AMBlock "github.com/ayachain/go-aya/vdb/mblock"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
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
	//TxCount					= TxHashIteratorLimit
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

	AtxThreadTxPackage 		AtxThreadsName = "thread.tx.package"

	AtxThreadExecutor		AtxThreadsName = "thread.block.executor"

	AtxThreadReceiptListen 	AtxThreadsName = "thread.receipt.listen"

	AtxThreadTxCommit 		AtxThreadsName = "thread.tx.commit"

	AtxThreadMining			AtxThreadsName = "thread.block.mining"

	AtxThreadTopicsListen	AtxThreadsName = "thread.topics.listen"

)

type ATxPool struct {

	cvfs vdb.CVFS
	storage *leveldb.DB
	ownerAccount EAccount.Account
	ownerAsset *AAssets.Assets

	adbpath string
	channelTopics string
	genBlock *ABlock.GenBlock

	topAssets []*AAssets.SortAssets
	workmode AtxPoolWorkMode
	ind *core.IpfsNode
	miningBlock *AMBlock.MBlock

	workctx context.Context
	workcancel context.CancelFunc

	threadChans map[AtxThreadsName] chan *AKeyStore.ASignedRawMsg
	notary ACore.Notary
	threadClosewg sync.WaitGroup
}


func NewTxPool( ind *core.IpfsNode, gblk *ABlock.GenBlock, cvfs vdb.CVFS, miner ACore.Notary, acc EAccount.Account) *ATxPool {

	adbpath := "/atxpool/" + gblk.ChainID
	var nd *merkledag.ProtoNode
	dsk := datastore.NewKey(adbpath)
	val, err := ind.Repo.Datastore().Get(dsk)

	// create channel topices string
	topic := crypto.Keccak256Hash( []byte( AtxPoolVersion + gblk.ChainID ) )

	switch {
	case err == datastore.ErrNotFound || val == nil:
		nd = unixfs.EmptyDirNode()

	case err == nil:

		c, err := cid.Cast(val)
		if err != nil {
			nd = unixfs.EmptyDirNode()
		}

		rnd, err := ind.DAG.Get(context.TODO(), c)
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

	oast, err := cvfs.Assetses().AssetsOf(acc.Address)
	if err != nil {
		oast = &AAssets.Assets{
			Version:AAssets.DRVer,
			Avail:0,
			Locked:0,
			Vote:0,
		}
	}

	root, err := mfs.NewRoot(
		context.TODO(),
		ind.DAG,
		nd,
		func(ctx context.Context, fcid cid.Cid) error {
			return ind.Repo.Datastore().Put(dsk, fcid.Bytes())
		},
	)

	dbnode, err := mfs.Lookup(root, "/" + gblk.ChainID)
	if err != nil {

		if err := mfs.Mkdir(root, "/" + gblk.ChainID, mfs.MkdirOpts{Flush:false, Mkparents:true}); err != nil {
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
			threadChans:make(map[AtxThreadsName] chan *AKeyStore.ASignedRawMsg),
			ownerAccount:acc,
			ownerAsset:oast,
			notary:miner,
			genBlock:gblk,
		}
	}

configNewDir :

	newnode, err := mfs.Lookup(root, "/" + gblk.ChainID)
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
		threadChans:make(map[AtxThreadsName] chan *AKeyStore.ASignedRawMsg),
		ownerAccount:acc,
		ownerAsset:oast,
		notary:miner,
		genBlock:gblk,
	}

}

func (pool *ATxPool) PowerOn( pctx context.Context ) error {

	pool.judgingMode()

	pool.workctx, pool.workcancel = context.WithCancel(pctx)

	fmt.Println("Waiting Thread Starting ...")

	switch pool.workmode {

	case AtxPoolWorkModeOblivioned:

		pool.runthreads(
			AtxThreadTopicsListen,
			AtxThreadTxCommit,
			AtxThreadTxPackage,
			AtxThreadMining,
			AtxThreadReceiptListen,
			AtxThreadExecutor,
		)

	case AtxPoolWorkModeSuper:

		pool.runthreads(
			AtxThreadTopicsListen,
			AtxThreadTxCommit,
			AtxThreadTxPackage,
			AtxThreadMining,
			AtxThreadReceiptListen,
			AtxThreadExecutor,
		)



	case AtxPoolWorkModeMaster:

		pool.runthreads(
			AtxThreadTopicsListen,
			AtxThreadTxCommit,
			//AtxThreadTxPackage,
			AtxThreadMining,
			//AtxThreadReceiptListen,
			AtxThreadExecutor,
		)


	case AtxPoolWorkModeNormal:

		pool.runthreads(
			AtxThreadTopicsListen,
			AtxThreadTxCommit,
			//AtxThreadTxPackage,
			//AtxThreadMining,
			//AtxThreadReceiptListen,
			AtxThreadExecutor,
		)

	}

	go func() {

		select {
		case <- pool.workctx.Done():

			pool.threadClosewg.Wait()

			for k, cc := range pool.threadChans {
				close(cc)
				delete(pool.threadChans, k)
			}

			if err := pool.cvfs.Close(); err != nil {
				log.Error(err)
			}

			if err := pool.storage.Close(); err != nil {
				log.Error(err)
			}

		}

	}()

	return nil
}

func (pool *ATxPool) PowerOff(err error) {

	if err != nil {
		log.Error(err)
	}

	pool.workcancel()
}

func (pool *ATxPool) UpdateBestBlock( cblock *ABlock.Block  ) error {

	idx, err := pool.cvfs.Indexes().GetIndex( cblock.Index )
	if err != nil {
		return err
	}

	if err := pool.cvfs.SeekToBlock( idx.FullCID ); err != nil {
		return err
	}

	ast, err := pool.cvfs.Assetses().AssetsOf( pool.ownerAccount.Address )

	if err != nil {
		return err
	}

	tops, err := pool.cvfs.Assetses().GetLockedTop100()
	if err != nil {
		tops = []*AAssets.SortAssets{}
	}

	pool.ownerAsset = ast
	pool.topAssets = tops

	return nil
}

func (pool *ATxPool) DoPackMBlock() {

	cc, exist := pool.threadChans[AtxThreadTxPackage]

	if !exist {
		return
	}

	cc <- nil

	return
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
func (pool *ATxPool) AddConfrimReceipt( mbhash EComm.Hash, retcid cid.Cid, from EComm.Address) {

	/// If it is not possible to obtain the proof of "VoteRight" from the source, it means
	/// that the result has no reference value and is discarded directly.
	ast, err := pool.cvfs.Assetses().AssetsOf( from )
	if err != nil {
		return
	}

	receiptKey := []byte(TxReceiptPrefix)
	copy(receiptKey[1:], mbhash.Bytes())

	exist, err := pool.storage.Has(receiptKey, nil)
	if err != nil {
		pool.PowerOff(ErrStorageLowAPIExpected)
	}

	rcidstr := retcid.String()
	receiptMap := make(map[string]uint64)
	if exist {

		value, err := pool.storage.Get(receiptKey, nil)
		if err != nil {
			pool.PowerOff(ErrStorageLowAPIExpected)
		}

		if json.Unmarshal(value, receiptMap) != nil {
			pool.PowerOff(ErrStorageRawDataDecodeExpected)
		}

		ocount, vexist := receiptMap[rcidstr]
		if vexist {
			receiptMap[rcidstr] = ocount + ast.Vote
		} else {
			receiptMap[rcidstr] = ast.Vote
		}

	} else {
		receiptMap[rcidstr] = ast.Vote
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
		pool.PowerOff(ErrStorageLowAPIExpected)
	}

	if err := pool.storage.Put( receiptKey, rmapbs, nil ); err != nil {
		pool.PowerOff(ErrStorageLowAPIExpected)
	}

}

func (pool *ATxPool) AddRawTransaction( tx *AKeyStore.ASignedRawMsg ) error {

	if !tx.Verify() {
		return ErrMessageVerifyExpected
	}

	subTx := ATx.Transaction{}
	if err := subTx.Decode(tx.Content[1:]); err != nil {
		return err
	}

	txraw := subTx.Encode()
	if len(txraw) <= 0 {
		return ErrMessageVerifyExpected
	}

	key := crypto.Keccak256(txraw)
	if err := pool.storage.Put(key, txraw, nil); err != nil {
		pool.PowerOff(err)
		return err
	}

	pool.DoPackMBlock()

	return nil
}

func (pool *ATxPool) ReadOnlyCVFS() vdb.CVFS {
	return pool.cvfs
}

/// private method
/// Judging working mode
func (pool *ATxPool) judgingMode() {

	pool.workmode = AtxPoolWorkModeOblivioned
	return

	n := 99999999

	for i := 0; i < len(pool.topAssets); i++ {
		if pool.topAssets[i] != nil && pool.topAssets[i].Address == pool.ownerAccount.Address {
			n = i
			break
		}
	}

	if n <= 21 {
		pool.workmode = AtxPoolWorkModeSuper
		return
	}

	oasset, err := pool.cvfs.Assetses().AssetsOf( pool.ownerAccount.Address )
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
			pool.PowerOff(err)
		}

		if err := dbtransaction.Write(batch, nil); err != nil {
			pool.PowerOff(err)
		}

		if err := dbtransaction.Commit(); err != nil {
			pool.PowerOff(err)
		}

	}

}

func (pool *ATxPool) sign( coder AvdbComm.AMessageEncode ) (*AKeyStore.ASignedRawMsg, error) {

	cbs := coder.RawMessageEncode()

	if len(cbs) <= 0 {
		return nil, ErrRawDBEndoedZeroLen
	}

	return AKeyStore.CreateMsg(cbs, pool.ownerAccount)
}

func (pool *ATxPool) runthreads( names ... AtxThreadsName ) {

	for _, n := range names {

		pool.threadChans[n] = make(chan *AKeyStore.ASignedRawMsg)

		switch n {

		case AtxThreadTopicsListen:
			go pool.channelListening(pool.workctx)


		case AtxThreadTxPackage:
			go pool.txPackageThread(pool.workctx)


		case AtxThreadReceiptListen:
			go pool.receiptListen(pool.workctx)


		case AtxThreadExecutor:
			go pool.blockExecutorThread(pool.workctx)


		case AtxThreadMining:
			go pool.miningThread(pool.workctx)

		}

	}

	time.Sleep(time.Microsecond * 100)

}
