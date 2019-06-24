package vdb

import (
	"context"
	"errors"
	"fmt"
	AWrok "github.com/ayachain/go-aya/consensus/core/worker"
	AAssetses "github.com/ayachain/go-aya/vdb/assets"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
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
	"github.com/syndtr/goleveldb/leveldb/opt"
	"sync"
)

var (
	ErrVDBServicesNotExist = errors.New("vdb services not exist in cvfs")
)

type CVFS interface {

	Close() error

	// In Data storage
	Indexes() AIndexes.IndexesServices

	// In IPFS DAG
	Blocks() ABlock.Services
	Assetses() AAssetses.Services
	Receipts() AReceipts.Services
	Transactions() ATx.Services
	Nodes() ANodes.Services

	WriteTaskGroup( group *AWrok.TaskBatchGroup ) ( cid.Cid, error )

	NewCVFSCache() (CacheCVFS, error)
}

type aCVFS struct {

	CVFS
	*mfs.Root

	inode *core.IpfsNode
	servies sync.Map
	indexServices AIndexes.IndexesServices
	chainId string
}

func CreateVFS( block *ABlock.GenBlock, ind *core.IpfsNode ) (cid.Cid, error) {

	var err error

	cvfs, err := LinkVFS(block.ChainID, cid.Undef, ind)

	if err != nil {
		return cid.Undef, err
	}
	defer cvfs.Close()

	writer, err := cvfs.NewCVFSCache()
	if err != nil {
		return cid.Undef, err
	}
	defer writer.Close()

	// Award
	for addr, amount := range block.Award {

		assetBn := AAssetses.NewAssets( amount, amount, 0 )

		writer.Assetses().PutNewAssets( EComm.HexToAddress(addr), assetBn )

	}

	// Block
	writer.Blocks().WriteGenBlock( block )

	group := writer.MergeGroup()

	// Group Write
	baseCid, err := cvfs.WriteTaskGroup(group)
	if err != nil {
		return cid.Undef, err
	}

	if err := cvfs.Indexes().PutIndexBy( 0, block.GetHash(), baseCid); err != nil {
		return cid.Undef, err
	}

	return baseCid, nil
}

func LinkVFS( chainId string, baseCid cid.Cid, ind *core.IpfsNode ) (CVFS, error) {

	root, err := newMFSRoot( context.TODO(), baseCid, ind )

	if err != nil {
		return nil, err
	}

	vfs := &aCVFS{
		Root:root,
		inode:ind,
		chainId:chainId,
	}

	vfs.indexServices = AIndexes.CreateServices(vfs.inode, vfs.chainId)

	/// ANodes VDBServices
	if dir, err := AVdbComm.LookupDBPath(vfs.Root, ANodes.DBPath); err != nil {
		goto experctedErr
	} else {
		vfs.servies.Store( ANodes.DBPath, ANodes.CreateServices(dir) )
	}

	/// AAssetses VDBServices
	if dir, err := AVdbComm.LookupDBPath(vfs.Root, AAssetses.DBPATH); err != nil {
		goto experctedErr
	} else {
		vfs.servies.Store( AAssetses.DBPATH, AAssetses.CreateServices(dir) )
	}

	/// ABlock VDBServices
	if dir, err := AVdbComm.LookupDBPath(vfs.Root, ABlock.DBPath); err != nil {
		goto experctedErr
	} else {
		vfs.servies.Store( ABlock.DBPath, ABlock.CreateServices(dir, vfs.indexServices) )
	}

	/// ATx VDBServices
	if dir, err := AVdbComm.LookupDBPath(vfs.Root, ATx.DBPath); err != nil {
		goto experctedErr
	} else {
		vfs.servies.Store( ATx.DBPath, ATx.CreateServices(dir) )
	}

	/// AReceipts VDBServices
	if dir, err := AVdbComm.LookupDBPath(vfs.Root, AReceipts.DBPath); err != nil {
		goto experctedErr
	} else {
		vfs.servies.Store( AReceipts.DBPath, AReceipts.CreateServices(dir) )
	}

	return vfs, nil

experctedErr:

	return nil, ErrVDBServicesNotExist
}

func ( vfs *aCVFS ) Nodes() ANodes.Services {

	v, _ := vfs.servies.Load(ANodes.DBPath )

	return v.(ANodes.Services)
}

func ( vfs *aCVFS ) Assetses() AAssetses.Services {

	v, _ := vfs.servies.Load(AAssetses.DBPATH)

	return v.(AAssetses.Services)
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

func ( vfs *aCVFS ) Indexes() AIndexes.IndexesServices {
	return vfs.indexServices
}

func ( vfs *aCVFS ) NewCVFSCache() (CacheCVFS, error) {
	return NewCacheCVFS(vfs)
}

func ( vfs *aCVFS ) WriteTaskGroup( group *AWrok.TaskBatchGroup) (cid.Cid, error) {

	var err error

	bmap := group.GetBatchMap()

	txs := &Transaction{
		transactions:make(map[string]*leveldb.Transaction),
	}

	for dbkey, batch := range bmap {

		services, exist := vfs.servies.Load(dbkey)
		if !exist {
			return cid.Undef, ErrVDBServicesNotExist
		}

		txs.transactions[dbkey], err = services.(AVdbComm.VDBSerices).OpenTransaction()
		if err != nil {
			return cid.Undef, err
		}

		if err := txs.transactions[dbkey].Write(batch, &opt.WriteOptions{Sync:true}); err != nil {
			return cid.Undef, err
		}
	}

	if err := txs.Commit(); err != nil {
		return cid.Undef, err
	}

	nd, err := mfs.FlushPath(context.TODO(), vfs.Root, "/")
	if err != nil {
		return cid.Undef, err
	}

	//Update
	for dbkey, _ := range bmap {

		vdbser, _ := vfs.servies.Load(dbkey)

		if err := vdbser.(AVdbComm.VDBSerices).UpdateSnapshot(); err != nil {
			log.Error(err)
		}
	}

	return nd.Cid(), nil
}

//func ( vfs *aCVFS ) Flush() cid.Cid {
//
//	nd, err := mfs.FlushPath(context.TODO(), vfs.Root, "/")
//	if err != nil {
//		return cid.Undef
//	}
//
//	return nd.Cid()
//}

func ( vfs *aCVFS ) Close() error {

	vfs.servies.Range(func(k,v interface{})bool{

		ser := v.(AVdbComm.VDBSerices)

		if err := ser.Shutdown(); err != nil {
			panic(err)
		}

		vfs.servies.Delete(k)

		return true

	})

	if err := vfs.indexServices.Close(); err != nil {
		return err
	}

	nd, err := mfs.FlushPath(context.TODO(), vfs.Root, "/")
	if err != nil {
		return err
	}

	defer fmt.Println("AfterClose MainCVFS CID:" + nd.Cid().String())

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

		pbnd, _ = rnd.(*merkledag.ProtoNode)
	}

	mroot, err := mfs.NewRoot(ctx, ind.DAG, pbnd, func(i context.Context, i2 cid.Cid) error {

		fmt.Println(i2.String())

		return nil
	})

	if err != nil {
		return nil, err
	}

	return mroot, nil
}
