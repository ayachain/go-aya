package vdb

import (
	"context"
	"errors"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AAssets "github.com/ayachain/go-aya/vdb/assets"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	VDBComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/im"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
	VDBMerge "github.com/ayachain/go-aya/vdb/merger"
	ANodes "github.com/ayachain/go-aya/vdb/node"
	AReceipts "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	"github.com/ipfs/go-unixfs"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

type CVFS interface {

	Indexes() AIndexes.IndexesServices

	Nodes() ANodes.Services

	Blocks() ABlock.Services

	Assetses() AAssets.Services

	Receipts() AReceipts.Services

	Transactions() ATx.Services

	ForkMergeBatch( merger VDBMerge.CVFSMerger) ( cid.Cid, error )

	NewCVFSWriter() (CacheCVFS, error)
}

type aCVFS struct {

	CVFS

	indexServices AIndexes.IndexesServices

	inode *core.IpfsNode
	chainId string
	servies sync.Map
	smu sync.Mutex
}

func CreateVFS( block *im.GenBlock, ind *core.IpfsNode, idxSer AIndexes.IndexesServices ) (cid.Cid, error) {

	if idxSer == nil {
		return cid.Undef, errors.New("indexes services can not be nil")
	}

	lidx, err := idxSer.GetLatest()
	if err != nil {
		//if err != leveldb.ErrNotFound {
		return cid.Undef, err
		//}
	}

	if lidx != nil {
		return lidx.FullCID, errors.New("chain indexes is already in repo")
	}

	/// gen this chain cvfs
	root, err := newMFSRoot( context.TODO(), cid.Undef, ind )
	if err != nil {
		return cid.Undef, err
	}
	defer func() {
		_ = root.Close()
	}()

	cvfs := &aCVFS{
		inode:ind,
		chainId:block.ChainID,
		indexServices:idxSer,
	}

	cvfs.servies.Store( ANodes.DBPath, ANodes.CreateServices(ind, idxSer) )
	cvfs.servies.Store( AAssets.DBPath, AAssets.CreateServices(ind, idxSer) )
	cvfs.servies.Store( ABlock.DBPath, ABlock.CreateServices(ind, idxSer) )
	cvfs.servies.Store( ATx.DBPath, ATx.CreateServices(ind, idxSer) )
	cvfs.servies.Store( AReceipts.DBPath, AReceipts.CreateServices(ind, idxSer) )

	writer, err := cvfs.NewCVFSWriter()
	defer func(){
		if err := writer.Close(); err != nil {
			log.Error(err)
		}
	}()

	if err != nil {
		return cid.Undef, err
	}

	// Award
	for _, gast := range block.Award {

		ast := AAssets.NewAssets( gast.Avail, gast.Vote, gast.Locked )

		writer.Assetses().Put( EComm.BytesToAddress( gast.Owner ), ast )
	}

	// Block
	writer.Blocks().WriteGenBlock( block )

	// Bootstrap Nodes
	writer.Nodes().InsertBootstrapNodes( block.SuperNodes )

	// Group Write
	baseCid, err := cvfs.writeGenBatch(root, writer.MergeGroup())
	if err != nil {
		return cid.Undef, err
	}

	if idxsCid, err := idxSer.PutIndexBy( 0, block.GetHash(), baseCid); err != nil {

		return cid.Undef, err

	} else {

		if err := idxSer.SyncToCID(idxsCid); err != nil {
			return cid.Undef, err
		}
	}

	return baseCid, nil
}

func LinkVFS( chainId string, ind *core.IpfsNode, idxSer AIndexes.IndexesServices ) (CVFS, error) {

	if idxSer == nil {
		return nil, errors.New("indexes services can not be nil")
	}

	lidx, err := idxSer.GetLatest()
	if err != nil || lidx == nil {
		return nil, errors.New("current indexes services doesn't look like an existing chain")
	}

	vfs := &aCVFS{
		inode:ind,
		chainId:chainId,
		indexServices:idxSer,
	}

	vfs.servies.Store( ANodes.DBPath, ANodes.CreateServices(ind, idxSer) )
	vfs.servies.Store( AAssets.DBPath, AAssets.CreateServices(ind, idxSer) )
	vfs.servies.Store( ABlock.DBPath, ABlock.CreateServices(ind, idxSer) )
	vfs.servies.Store( ATx.DBPath, ATx.CreateServices(ind, idxSer) )
	vfs.servies.Store( AReceipts.DBPath, AReceipts.CreateServices(ind, idxSer) )

	return vfs, nil
}

func ( vfs *aCVFS ) Indexes() AIndexes.IndexesServices {

	return vfs.indexServices
}

func ( vfs *aCVFS ) Nodes() ANodes.Services {

	v, _ := vfs.servies.Load(ANodes.DBPath )
	return v.(ANodes.Services)
}

func ( vfs *aCVFS ) Assetses() AAssets.Services {

	v, _ := vfs.servies.Load(AAssets.DBPath)

	return v.(AAssets.Services)
}

func ( vfs *aCVFS ) Blocks() ABlock.Services {

	v, _ := vfs.servies.Load(ABlock.DBPath)

	return v.(ABlock.Services)
}

func ( vfs *aCVFS ) Transactions() ATx.Services {

	v, _ := vfs.servies.Load(ATx.DBPath)

	return v.(ATx.Services)
}

func ( vfs *aCVFS ) Receipts() AReceipts.Services {

	v, _ := vfs.servies.Load(AReceipts.DBPath)

	return v.(AReceipts.Services)
}

func ( vfs *aCVFS ) NewCVFSWriter() (CacheCVFS, error) {
	return NewCacheCVFS(vfs)
}

func ( vfs *aCVFS ) writeGenBatch( root *mfs.Root, merger VDBMerge.CVFSMerger ) (cid. Cid, error) {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	if err := merger.ForEach( func(k string, bc *leveldb.Batch ) error {

		vdbroot, err := VDBComm.LookupDBPath(root, k)
		if err != nil {
			panic(err)
		}

		if err := ADB.MergeClose(vdbroot, bc, k); err != nil {
			return err
		}

		return nil

	}); err != nil {

		return cid.Undef, err

	}

	nd, err := mfs.FlushPath(context.TODO(), root, "/")
	if err != nil {
		return cid.Undef, err
	}

	return nd.Cid(), nil
}

func ( vfs *aCVFS ) ForkMergeBatch( merger VDBMerge.CVFSMerger ) (cid.Cid, error) {

	var (
		lidx *AIndexes.Index
		err error
	)

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	var root *mfs.Root

	if lidx, err = vfs.indexServices.GetLatest(); err != nil {

		panic(err)

	} else if lidx == nil {

		panic(errors.New("get latest idx expected"))

	} else {

		if rt, err := newMFSRoot( context.TODO(), lidx.FullCID, vfs.inode ); err != nil {
			panic(err)
		} else {
			root = rt
		}
	}

	if err := merger.ForEach(func(k string, bc *leveldb.Batch) error {

		vdbroot, err := VDBComm.LookupDBPath(root, k)
		if err != nil {
			panic(err)
		}

		if err := ADB.MergeClose(vdbroot, bc, k); err != nil {
			return err
		}

		return nil

	}); err != nil {
		return cid.Undef, err
	}
	
	nd, err := mfs.FlushPath(context.TODO(), root, "/")
	if err != nil {
		return cid.Undef, err
	}

	return nd.Cid(), nil
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

		pbnd, _ = rnd.(*merkledag.ProtoNode)
	}

	mroot, err := mfs.NewRoot(ctx, ind.DAG, pbnd, func(i context.Context, i2 cid.Cid) error {
		return nil
	})

	if err != nil {
		return nil, err
	}

	return mroot, nil
}
