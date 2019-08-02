package vdb

import (
	"context"
	"errors"
	AWrok "github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ayachain/go-aya/logs"
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
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/whyrusleeping/go-logging"
	"sync"
)

var (
	ErrVDBServicesNotExist = errors.New("vdb services not exist in cvfs")
)

var log = logging.MustGetLogger(logs.AModules_CVFS)

type CVFS interface {

	Close() error

	// In Data storage
	Indexes() AIndexes.IndexesServices

	// In IPFS DAG
	Nodes() ANodes.Services
	Blocks() ABlock.Services
	Assetses() AAssetses.Services
	Receipts() AReceipts.Services
	Transactions() ATx.Services

	Restart( baseCid cid.Cid ) error
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
		log.Error(err)
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

		assetBn := AAssetses.NewAssets( amount / 2, amount / 2, amount )

		writer.Assetses().PutNewAssets( EComm.HexToAddress(addr), assetBn )

	}

	// Block
	writer.Blocks().WriteGenBlock( block )

	// Bootstrap Nodes
	writer.Nodes().InsertBootstrapNodes( block.SuperNodes )

	group := writer.MergeGroup()

	// Group Write
	baseCid, err := cvfs.WriteTaskGroup(group)
	if err != nil {
		return cid.Undef, err
	}

	if err := cvfs.Indexes().PutIndexBy( 0, block.GetHash(), baseCid); err != nil {
		return cid.Undef, err
	}

	log.Debugf("CID:%v", baseCid.String())

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

	if err := vfs.initServices(); err != nil {
		return nil, err
	} else {
		return vfs, nil
	}
}

func ( vfs *aCVFS ) Restart( baseCid cid.Cid ) error {

	if err := vfs.Close(); err != nil {
		return err
	}

	root, err := newMFSRoot( context.TODO(), baseCid, vfs.inode )

	if err != nil {
		return err
	}

	vfs.Root = root

	return vfs.initServices()
}

func ( vfs *aCVFS ) Nodes() ANodes.Services {

	v, _ := vfs.servies.Load(ANodes.DBPath )

	return v.(ANodes.Services)
}

func ( vfs *aCVFS ) Assetses() AAssetses.Services {

	v, _ := vfs.servies.Load(AAssetses.DBPath)

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

	// Flush VDB
	vfs.servies.Range(func(key, value interface{}) bool {

		if err := value.(AVdbComm.VDBSerices).SyncCache(); err != nil {
			log.Error(err)
			return false
		}

		return true

	})

	nd, err := mfs.FlushPath(context.TODO(), vfs.Root, "/")
	if err != nil {
		return cid.Undef, err
	}

	//Update Snapshot
	for dbkey, _ := range bmap {

		vdbser, _ := vfs.servies.Load(dbkey)

		if err := vdbser.(AVdbComm.VDBSerices).UpdateSnapshot(); err != nil {
			log.Error(err)
		}
	}

	return nd.Cid(), nil
}

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

	_, err := mfs.FlushPath(context.TODO(), vfs.Root, "/")
	if err != nil {
		return err
	}

	return vfs.Root.Close()
}

func ( vfs *aCVFS ) initServices() error {

	vfs.indexServices = AIndexes.CreateServices(vfs.inode, vfs.chainId)

	/// ANodes VDBServices
	if dir, err := AVdbComm.LookupDBPath(vfs.Root, ANodes.DBPath); err != nil {
		goto experctedErr
	} else {
		vfs.servies.Store( ANodes.DBPath, ANodes.CreateServices(dir) )
	}

	/// AAssetses VDBServices
	if dir, err := AVdbComm.LookupDBPath(vfs.Root, AAssetses.DBPath); err != nil {
		goto experctedErr
	} else {
		vfs.servies.Store( AAssetses.DBPath, AAssetses.CreateServices(dir) )
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

	return nil

experctedErr:

	return ErrVDBServicesNotExist
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
