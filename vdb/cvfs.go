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
	"strings"
	"sync"
)

var (
	ErrVDBServicesNotExist = errors.New("vdb services not exist in cvfs")
)

var log = logging.MustGetLogger(logs.AModules_CVFS)

type CVFS interface {

	Close() error

	Indexes() AIndexes.IndexesServices

	Nodes() ANodes.Services

	Blocks() ABlock.Services

	Assetses() AAssetses.Services

	Receipts() AReceipts.Services

	Transactions() ATx.Services

	Restart( baseCid cid.Cid ) error

	WriteTaskGroup( group *AWrok.TaskBatchGroup ) ( cid.Cid, error )

	NewCVFSCache() (CacheCVFS, error)

	BestCID() cid.Cid
}

type aCVFS struct {

	CVFS
	*mfs.Root

	chainId string
	bestCID cid.Cid

	servies sync.Map
	smu sync.Mutex
	inode *core.IpfsNode

	indexServices AIndexes.IndexesServices
}

func CreateVFS( block *ABlock.GenBlock, ind *core.IpfsNode, idxSer AIndexes.IndexesServices ) (cid.Cid, error) {

	if idxSer == nil {
		return cid.Undef, errors.New("indexes services can not be nil")
	}

	lidx, err := idxSer.GetLatest()
	if err != nil {

		if err != leveldb.ErrNotFound {
			return cid.Undef, err
		}

	}

	if lidx != nil {
		return lidx.FullCID, errors.New("chain indexes is already in repo")
	}

	/// gen this chain cvfs
	root, err := newMFSRoot( context.TODO(), cid.Undef, ind )
	if err != nil {
		return cid.Undef, err
	}

	cvfs := &aCVFS{
		Root:root,
		inode:ind,
		chainId:block.ChainID,
		bestCID:cid.Undef,
		indexServices:idxSer,
	}
	defer func () {
		if  cvfs != nil {
			if err := cvfs.Close(); err != nil {
				log.Error(err)
			}
		}
	}()

	if err := cvfs.initServices(); err != nil {
		return cid.Undef, err
	}

	writer, err := cvfs.NewCVFSCache()
	defer func(){
		if err := writer.Close(); err != nil {
			log.Error(err)
		}
	}()

	if err != nil {
		return cid.Undef, err
	}

	// Award
	for addr, amount := range block.Award {

		assetBn := AAssetses.NewAssets( amount / 2, amount, amount / 2 )

		writer.Assetses().PutNewAssets( EComm.HexToAddress(addr), assetBn )

	}

	// Block
	writer.Blocks().WriteGenBlock( block )

	// Bootstrap Nodes
	writer.Nodes().InsertBootstrapNodes( block.SuperNodes )

	// Group Write
	baseCid, err := cvfs.WriteTaskGroup(writer.MergeGroup())
	if err != nil {
		return cid.Undef, err
	}

	if err := idxSer.PutIndexBy( 0, block.GetHash(), baseCid); err != nil {
		return cid.Undef, err
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

	root, err := newMFSRoot( context.TODO(), lidx.FullCID, ind )

	if err != nil {
		return nil, err
	}

	vfs := &aCVFS{
		Root:root,
		inode:ind,
		chainId:chainId,
		bestCID:lidx.FullCID,
		indexServices:idxSer,
	}

	if err := vfs.initServices(); err != nil {
		return nil, err
	} else {
		return vfs, nil
	}
}

func ( vfs *aCVFS ) BestCID() cid.Cid {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	return vfs.bestCID

}

func ( vfs *aCVFS ) Restart( baseCid cid.Cid ) error {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	if strings.EqualFold(baseCid.String(), vfs.bestCID.String()) {
		return nil
	}

	if err := vfs.Close(); err != nil {
		return err
	}

	root, err := newMFSRoot( context.TODO(), baseCid, vfs.inode )

	if err != nil {
		return err
	}

	vfs.Root = root
	vfs.bestCID = baseCid

	return vfs.initServices()
}

func ( vfs *aCVFS ) Indexes() AIndexes.IndexesServices {
	return vfs.indexServices
}

func ( vfs *aCVFS ) Nodes() ANodes.Services {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	v, _ := vfs.servies.Load(ANodes.DBPath )

	return v.(ANodes.Services)
}

func ( vfs *aCVFS ) Assetses() AAssetses.Services {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	v, _ := vfs.servies.Load(AAssetses.DBPath)

	return v.(AAssetses.Services)
}

func ( vfs *aCVFS ) Blocks() ABlock.Services {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	v, _ := vfs.servies.Load(ABlock.DBPath)

	return v.(ABlock.Services)
}

func ( vfs *aCVFS ) Transactions() ATx.Services {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	v, _ := vfs.servies.Load(ATx.DBPath)

	return v.(ATx.Services)
}

func ( vfs *aCVFS ) Receipts() AReceipts.Services {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	v, _ := vfs.servies.Load(AReceipts.DBPath)
	return v.(AReceipts.Services)
}

func ( vfs *aCVFS ) NewCVFSCache() (CacheCVFS, error) {
	return NewCacheCVFS(vfs)
}

func ( vfs *aCVFS ) WriteTaskGroup( group *AWrok.TaskBatchGroup) (cid.Cid, error) {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	var err error

	bmap := group.GetBatchMap()

	txs := &Transaction{
		transactions:make(map[string]*leveldb.Transaction),
	}

	for dbkey, batch := range bmap {

		if batch == nil {
			continue
		}

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
	for dbkey := range bmap {

		vdbser, _ := vfs.servies.Load(dbkey)

		if err := vdbser.(AVdbComm.VDBSerices).UpdateSnapshot(); err != nil {
			log.Error(err)
		}
	}

	return nd.Cid(), nil
}

func ( vfs *aCVFS ) Close() error {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

	vfs.servies.Range(func(k,v interface{})bool{

		ser := v.(AVdbComm.VDBSerices)

		if err := ser.Shutdown(); err != nil {
			panic(err)
		}

		vfs.servies.Delete(k)

		return true

	})

	_, err := mfs.FlushPath(context.TODO(), vfs.Root, "/")
	if err != nil {
		return err
	}

	return vfs.Root.Close()
}

func ( vfs *aCVFS ) initServices() error {

	vfs.smu.Lock()
	defer vfs.smu.Unlock()

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
