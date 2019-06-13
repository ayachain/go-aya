package vdb

import (
	"context"
	"errors"
	AWrok "github.com/ayachain/go-aya/consensus/core/worker"
	AAssetses "github.com/ayachain/go-aya/vdb/assets"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
	AReceipts "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	"github.com/ipfs/go-unixfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"sync"
)

var (
	aVFSDAGNodeConversionError = errors.New("conversion proto node expected")
)

type CVFS interface {

	Close()						error
	SeekToBlock( bcid cid.Cid )  error

	Blocks() 					ABlock.BlocksAPI		/// Body
	Receipts() 					AReceipts.ReceiptsAPI	/// Receipt
	Assetses() 					AAssetses.AssetsAPI		/// Asset
	Transactions() 				ATx.TransactionAPI		/// Transaction
	Indexes()					AIndexes.IndexesAPI

	WriteTaskGroup( group *AWrok.TaskBatchGroup) (cid.Cid, error)

}

type aCVFS struct {

	CVFS
	*mfs.Root

	inode *core.IpfsNode

	ctx context.Context
	ctxCancel context.CancelFunc

	servies map[string]AVdbComm.VDBSerices
	indexServices AIndexes.IndexesAPI
	chainId string

	rwmutex sync.RWMutex
}

func CreateVFS( block *ABlock.GenBlock, ind *core.IpfsNode ) (cid.Cid, error) {

	var err error

	cvfs, err := LinkVFS(block.ChainID, cid.Undef, ind)
	if err != nil {
		return cid.Undef, err
	}
	defer cvfs.Close()

	genBatchGroup := AWrok.NewGroup()

	// Award
	for addr, amount := range block.Award {
		assetBn := AAssetses.NewAssets( amount, amount, 0 ).Encode()
		genBatchGroup.Put( cvfs.Assetses().DBKey(), EComm.HexToAddress(addr).Bytes(), assetBn )
	}

	// Block
	err = cvfs.Blocks().WriteGenBlock( genBatchGroup, block )
	if err != nil {
		return cid.Undef, err
	}

	// Group Write
	baseCid, err := cvfs.WriteTaskGroup(genBatchGroup)
	if err != nil {
		return cid.Undef, err
	}

	if err := cvfs.Indexes().PutIndexBy( 0, block.GetHash(), baseCid); err != nil {
		return cid.Undef, err
	}

	_ = cvfs.Close()

	return baseCid, nil
}


func LinkVFS( chainId string, baseCid cid.Cid, ind *core.IpfsNode ) (CVFS, error) {

	ctx, cancel := context.WithCancel( context.Background() )

	root, err := newMFSRoot( ctx, baseCid, ind )

	if err != nil {
		return nil, err
	}

	vfs := &aCVFS{
		ctx:ctx,
		ctxCancel:cancel,
		Root:root,
		inode:ind,
		chainId:chainId,
		servies: map[string]AVdbComm.VDBSerices{
		},
	}

	vfs.indexServices = AIndexes.CreateServices(vfs.inode, vfs.chainId)

	return vfs, nil
}

func ( vfs *aCVFS ) SeekToBlock( bcid cid.Cid ) error {

	if err := vfs.Root.Close(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	root, err := newMFSRoot(ctx, bcid, vfs.inode)

	if err != nil {
		return err
	}

	vfs.Root = root
	vfs.ctxCancel = cancel
	vfs.ctx = ctx

	return nil
}

func ( vfs *aCVFS ) Assetses() AAssetses.AssetsAPI {

	vfs.rwmutex.Lock()
	defer vfs.rwmutex.Unlock()

	v, exist := vfs.servies[ AAssetses.DBPATH ]

	if !exist {

		astDir, err := AVdbComm.LookupDBPath(vfs.Root, AAssetses.DBPATH)

		if err != nil {
			return nil
		}

		v = AAssetses.CreateServices(astDir,true)

		vfs.servies[ AAssetses.DBPATH ] = v
	}

	return v.(AAssetses.AssetsAPI)

}

func ( vfs *aCVFS ) Blocks() ABlock.BlocksAPI {

	vfs.rwmutex.Lock()
	defer vfs.rwmutex.Unlock()

	v, exist := vfs.servies[ ABlock.DBPath ]

	if !exist {

		blockDir, err := AVdbComm.LookupDBPath(vfs.Root, ABlock.DBPath)

		if err != nil {
			return nil
		}

		v = ABlock.CreateServices(blockDir, vfs.indexServices, true)

		vfs.servies[ ABlock.DBPath ] = v

	}

	return v.(ABlock.BlocksAPI)
}

func ( vfs *aCVFS ) Transactions() ATx.TransactionAPI {

	vfs.rwmutex.Lock()
	defer vfs.rwmutex.Unlock()

	v, exist := vfs.servies[ ATx.DBPath ]

	if !exist {

		atxDir, err := AVdbComm.LookupDBPath(vfs.Root, ATx.DBPath)

		if err != nil {
			return nil
		}

		v =  ATx.CreateServices(atxDir, true)

		vfs.servies[ ATx.DBPath ] = v

	}

	return v.(ATx.TransactionAPI)
}

func ( vfs *aCVFS ) Receipts() AReceipts.ReceiptsAPI {

	vfs.rwmutex.Lock()
	defer vfs.rwmutex.Unlock()

	v, exist := vfs.servies[ AReceipts.DBPath ]

	if !exist {

		rcpDir, err := AVdbComm.LookupDBPath(vfs.Root, AReceipts.DBPath)

		if err != nil {
			return nil
		}

		v =  AReceipts.CreateServices(rcpDir, true)

		vfs.servies[ AReceipts.DBPath ] = v

	}

	return v.(AReceipts.ReceiptsAPI)
}

func ( vfs *aCVFS ) Indexes() AIndexes.IndexesAPI {

	vfs.rwmutex.Lock()
	defer vfs.rwmutex.Unlock()

	return vfs.indexServices
}

func ( vfs *aCVFS ) WriteTaskGroup( group *AWrok.TaskBatchGroup) (cid.Cid, error) {

	vfs.rwmutex.RLock()
	defer vfs.rwmutex.RUnlock()

	for k, v := range vfs.servies {
		v.Close()
		delete(vfs.servies, k)
	}

	var err error

	bmap := group.GetBatchMap()

	txs := &Transaction{
		transactions:make(map[string]*leveldb.Transaction),
		lockers: make(map[string]*sync.RWMutex),
	}

	for dbkey, batch := range bmap {

		dir ,err := AVdbComm.LookupDBPath(vfs.Root, dbkey)

		if err != nil {
			return cid.Undef, err
		}

		var services AVdbComm.VDBSerices

		switch dbkey {

		case AAssetses.DBPATH:
			services = AAssetses.CreateServices( dir, false )

		case ABlock.DBPath:
			services = ABlock.CreateServices(dir, vfs.indexServices, false)

		case ATx.DBPath:
			services = ATx.CreateServices(dir, false)

		case AReceipts.DBPath:
			services = AReceipts.CreateServices(dir, false)
		}

		txs.transactions[dbkey], txs.lockers[dbkey], err = services.OpenVDBTransaction()
		if err != nil {
			return cid.Undef, nil
		}

		if err := txs.transactions[dbkey].Write(batch, &opt.WriteOptions{Sync:true}); err != nil {
			return cid.Undef, nil
		}

	}

	if err := txs.Commit(); err != nil {
		return cid.Undef, err
	}

	nd, err := mfs.FlushPath(context.TODO(), vfs.Root, "/")
	if err != nil {
		return cid.Undef, err
	}

	return nd.Cid(), nil
}

func ( vfs *aCVFS ) Close() error {

	if err := vfs.indexServices.Close(); err != nil {
		return err
	}

	if err := vfs.indexServices.Close(); err != nil {
		return err
	}

	return vfs.Root.Close()
}

func newMFSRoot( ctx context.Context, c cid.Cid, ind *core.IpfsNode ) ( *mfs.Root, error ) {

	var pbnd *merkledag.ProtoNode

	if c == cid.Undef {

		pbnd = unixfs.EmptyDirNode()

	} else {

		rnd, err := ind.DAG.Get(ctx, c)
		if err != nil {
			return nil, err
		}

		var ok bool
		pbnd, ok = rnd.(*merkledag.ProtoNode)
		if !ok {
			return nil, aVFSDAGNodeConversionError
		}

	}

	mroot, err := mfs.NewRoot(ctx, ind.DAG, pbnd, func(i context.Context, i2 cid.Cid) error {
		//fmt.Println("CVFSPublishedCID : " + i2.String())
		return nil
	})

	if err != nil {
		return nil, err
	}

	return mroot, nil
}
