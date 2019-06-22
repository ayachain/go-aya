package txpool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	ACore "github.com/ayachain/go-aya/consensus/core"
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
	"github.com/whyrusleeping/go-logging"
	"sync"
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

type ATxPoolThreadsName string

const (

	PackageTxsLimit 								= 2048

	AtxPoolWorkModeNormal 			AtxPoolWorkMode = 0

	AtxPoolWorkModeMaster 			AtxPoolWorkMode = 1

	AtxPoolWorkModeSuper 			AtxPoolWorkMode = 2

	AtxPoolWorkModeOblivioned 		AtxPoolWorkMode = 3


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

	ATxPoolThreadChainInfo			ATxPoolThreadsName 	= "thread.chain.info"
	ATxPoolThreadChainInfoBuff							= 1
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
}

func NewTxPool( ind *core.IpfsNode, gblk *ABlock.GenBlock, cvfs vdb.CVFS, miner ACore.Notary, acc EAccount.Account) *ATxPool {

	adbpath := "/atxpool/" + gblk.ChainID
	var nd *merkledag.ProtoNode
	dsk := datastore.NewKey(adbpath)
	val, err := ind.Repo.Datastore().Get(dsk)

	// create channel topices string
	topic := crypto.Keccak256Hash( []byte( AtxPoolVersion + gblk.ChainID ) )

	topicmap := map[ATxPoolThreadsName]string{
		ATxPoolThreadTxListen : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxListen))).String(),
		ATxPoolThreadTxPackage : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadTxPackage))).String(),
		ATxPoolThreadExecutor : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadExecutor))).String(),
		ATxPoolThreadReceiptListen : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadReceiptListen))).String(),
		ATxPoolThreadMining : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, ATxPoolThreadMining))).String(),
	}

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

		if err := mfs.Mkdir(root, "/" + gblk.ChainID, mfs.MkdirOpts{Flush:true, Mkparents:true}); err != nil {
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
			workmode:AtxPoolWorkModeNormal,
			ind:ind,
			channelTopics:topicmap,
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
		workmode:AtxPoolWorkModeNormal,
		ind:ind,
		channelTopics:topicmap,
		ownerAccount:acc,
		ownerAsset:oast,
		notary:miner,
		genBlock:gblk,
	}

}

func (pool *ATxPool) PowerOn( ctx context.Context ) error {

	pool.judgingMode()

	fmt.Println("ATxPool Working Started.")

	defer fmt.Println("ATxPool Working Stoped.")

	switch pool.workmode {

	case AtxPoolWorkModeOblivioned:

		pool.runThreads(
			ctx,
			ATxPoolThreadTxListen,
			ATxPoolThreadTxPackage,
			ATxPoolThreadMining,
			ATxPoolThreadReceiptListen,
			ATxPoolThreadExecutor,
		)

	case AtxPoolWorkModeSuper:

		pool.runThreads(
			ctx,
			ATxPoolThreadTxPackage,
			ATxPoolThreadMining,
			ATxPoolThreadReceiptListen,
			ATxPoolThreadExecutor,
		)



	case AtxPoolWorkModeMaster:

		pool.runThreads(
			ctx,
			//AtxThreadTxPackage,
			ATxPoolThreadMining,
			//AtxThreadReceiptListen,
			ATxPoolThreadExecutor,
		)


	case AtxPoolWorkModeNormal:

		pool.runThreads(
			ctx,
			//AtxThreadTxPackage,
			//AtxThreadMining,
			//AtxThreadReceiptListen,
			ATxPoolThreadExecutor,
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

func (pool *ATxPool) AddConfrimReceipt( mbhash EComm.Hash, retcid cid.Cid, from EComm.Address) error {

	/// If it is not possible to obtain the proof of "VoteRight" from the source, it means
	/// that the result has no reference value and is discarded directly.
	ast, err := pool.cvfs.Assetses().AssetsOf( from )
	if err != nil {
		return err
	}

	receiptKey := []byte(TxReceiptPrefix)
	copy(receiptKey[1:], mbhash.Bytes())

	exist, err := pool.storage.Has(receiptKey, nil)
	if err != nil {
		return ErrStorageLowAPIExpected
	}

	rcidstr := retcid.String()
	receiptMap := make(map[string]uint64)
	if exist {

		value, err := pool.storage.Get(receiptKey, nil)
		if err != nil {
			return ErrStorageLowAPIExpected
		}

		if json.Unmarshal(value, receiptMap) != nil {
			return ErrStorageRawDataDecodeExpected
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

	rmapbs, err := json.Marshal(receiptMap)
	if err != nil {
		return ErrStorageLowAPIExpected
	}

	if err := pool.storage.Put( receiptKey, rmapbs, nil ); err != nil {
		return ErrStorageLowAPIExpected
	}

	return nil
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

/// Private method
/// Judging working mode
func (pool *ATxPool) judgingMode() {

	pool.workmode = AtxPoolWorkModeOblivioned
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

	pool.DoPackMBlock()

	return nil
}