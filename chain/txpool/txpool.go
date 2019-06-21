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

	AtxThreadTxListen		AtxThreadsName = "thread.tx.listen"

	AtxThreadTxPackage 		AtxThreadsName = "thread.tx.package"

	AtxThreadExecutor		AtxThreadsName = "thread.block.executor"

	AtxThreadReceiptListen 	AtxThreadsName = "thread.receipt.listen"

	AtxThreadMining			AtxThreadsName = "thread.block.mining"

	AtxThreadChainInfo		AtxThreadsName = "thread.chain.info"

)

type ATxPool struct {

	cvfs vdb.CVFS
	storage *leveldb.DB
	ownerAccount EAccount.Account
	ownerAsset *AAssets.Assets

	adbpath string

	channelTopics map[AtxThreadsName]string

	genBlock *ABlock.GenBlock

	topAssets []*AAssets.SortAssets
	workmode AtxPoolWorkMode
	ind *core.IpfsNode
	miningBlock *AMBlock.MBlock

	threadChans map[AtxThreadsName] chan *AKeyStore.ASignedRawMsg
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

	topicmap := map[AtxThreadsName]string{
		AtxThreadTxListen : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, AtxThreadTxListen))).String(),
		AtxThreadTxPackage : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, AtxThreadTxPackage))).String(),
		AtxThreadExecutor : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, AtxThreadExecutor))).String(),
		AtxThreadReceiptListen : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, AtxThreadReceiptListen))).String(),
		AtxThreadMining : crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", topic, AtxThreadMining))).String(),
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
			channelTopics:topicmap,
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
		channelTopics:topicmap,
		threadChans:make(map[AtxThreadsName] chan *AKeyStore.ASignedRawMsg),
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
			AtxThreadTxListen,
			AtxThreadTxPackage,
			AtxThreadMining,
			AtxThreadReceiptListen,
			AtxThreadExecutor,
		)

	case AtxPoolWorkModeSuper:

		pool.runThreads(
			ctx,
			AtxThreadTxPackage,
			AtxThreadMining,
			AtxThreadReceiptListen,
			AtxThreadExecutor,
		)



	case AtxPoolWorkModeMaster:

		pool.runThreads(
			ctx,
			//AtxThreadTxPackage,
			AtxThreadMining,
			//AtxThreadReceiptListen,
			AtxThreadExecutor,
		)


	case AtxPoolWorkModeNormal:

		pool.runThreads(
			ctx,
			//AtxThreadTxPackage,
			//AtxThreadMining,
			//AtxThreadReceiptListen,
			AtxThreadExecutor,
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

	return pool.doBroadcast(tx, pool.channelTopics[AtxThreadTxListen])
}

/// Private method
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

func (pool *ATxPool) sign( coder AvdbComm.AMessageEncode ) (*AKeyStore.ASignedRawMsg, error) {

	cbs := coder.RawMessageEncode()

	if len(cbs) <= 0 {
		return nil, ErrRawDBEndoedZeroLen
	}

	return AKeyStore.CreateMsg(cbs, pool.ownerAccount)
}

func (pool *ATxPool) runThreads( ctx context.Context, names ... AtxThreadsName ) {

	for _, n := range names {

		sctx := context.WithValue( ctx, "Pool", pool )

		switch n {

		case AtxThreadTxListen:
			go txListenThread(sctx)


		case AtxThreadTxPackage:
			go txPackageThread(sctx)


		case AtxThreadMining:
			go miningThread(sctx)


		case AtxThreadReceiptListen:
			go receiptListen(sctx)


		case AtxThreadExecutor:
			go blockExecutorThread(sctx)

		}

	}

}

func (pool *ATxPool) doBroadcast( coder AvdbComm.AMessageEncode, topic string) error {

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


		return pool.ind.PubSub.Publish( topic, rawsmsg )
	}

}

func (pool *ATxPool) addRawTransaction( tx *AKeyStore.ASignedRawMsg ) error {

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
		return err
	}

	pool.DoPackMBlock()

	return nil
}
