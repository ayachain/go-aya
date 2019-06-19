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
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"sync"
)

var (
	aVFSDAGNodeConversionError = errors.New("conversion proto node expected")
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

	SeekToBlock( bcid cid.Cid ) error
	WriteTaskGroup( group *AWrok.TaskBatchGroup ) ( cid.Cid, error )

	NewCVFSCache() (CacheCVFS, error)
}

type aCVFS struct {

	CVFS
	*mfs.Root

	inode *core.IpfsNode

	servies map[string]AVdbComm.VDBSerices

	indexServices AIndexes.IndexesServices
	chainId string

	writeWaiter *sync.WaitGroup
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

	_ = writer.Close()
	_ = cvfs.Close()

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
		servies: make(map[string]AVdbComm.VDBSerices),
		writeWaiter: &sync.WaitGroup{},
	}

	vfs.indexServices = AIndexes.CreateServices(vfs.inode, vfs.chainId)

	return vfs, nil
}

func ( vfs *aCVFS ) SeekToBlock( bcid cid.Cid ) error {

	vfs.writeWaiter.Wait()
	vfs.writeWaiter.Add(1)
	defer vfs.writeWaiter.Done()

	lcid, err := mfs.FlushPath(context.TODO(), vfs.Root, "/")
	if err != nil {
		return err
	}

	if lcid.Cid().Equals( bcid ) {
		return nil
	}

	for k, v := range vfs.servies {
		if err := v.Shutdown(); err != nil {
			panic(err)
		}
		delete(vfs.servies, k)
	}

	if err := vfs.Root.Close(); err != nil {
		return err
	}

	root, err := newMFSRoot(context.TODO(), bcid, vfs.inode)

	if err != nil {
		return err
	}

	vfs.Root = root

	return nil
}

func ( vfs *aCVFS ) Assetses() AAssetses.Services {

	vfs.writeWaiter.Wait()

	v, exist := vfs.servies[ AAssetses.DBPATH ]

	if !exist {

		astDir, err := AVdbComm.LookupDBPath(vfs.Root, AAssetses.DBPATH)

		if err != nil {
			return nil
		}

		v = AAssetses.CreateServices(astDir,true)

		vfs.servies[ AAssetses.DBPATH ] = v
	}

	return v.(AAssetses.Services)

}

func ( vfs *aCVFS ) Blocks() ABlock.Services {

	vfs.writeWaiter.Wait()

	v, exist := vfs.servies[ ABlock.DBPath ]

	if !exist {

		blockDir, err := AVdbComm.LookupDBPath(vfs.Root, ABlock.DBPath)

		if err != nil {
			return nil
		}

		v = ABlock.CreateServices(blockDir, vfs.indexServices, true)

		vfs.servies[ ABlock.DBPath ] = v

	}

	return v.(ABlock.Services)
}

func ( vfs *aCVFS ) Transactions() ATx.Services {

	vfs.writeWaiter.Wait()

	v, exist := vfs.servies[ ATx.DBPath ]

	if !exist {

		atxDir, err := AVdbComm.LookupDBPath(vfs.Root, ATx.DBPath)

		if err != nil {
			return nil
		}


		v =  ATx.CreateServices(atxDir, true)

		vfs.servies[ ATx.DBPath ] = v

	}

	return v.(ATx.Services)
}

func ( vfs *aCVFS ) Receipts() AReceipts.Services {

	vfs.writeWaiter.Wait()

	v, exist := vfs.servies[ AReceipts.DBPath ]

	if !exist {

		rcpDir, err := AVdbComm.LookupDBPath(vfs.Root, AReceipts.DBPath)

		if err != nil {
			return nil
		}

		v =  AReceipts.CreateServices(rcpDir, true)

		vfs.servies[ AReceipts.DBPath ] = v

	}

	return v.(AReceipts.Services)
}

func ( vfs *aCVFS ) Indexes() AIndexes.IndexesServices {

	vfs.writeWaiter.Wait()

	return vfs.indexServices
}

func ( vfs *aCVFS ) NewCVFSCache() (CacheCVFS, error) {

	return NewCacheCVFS(vfs)

}

func ( vfs *aCVFS ) WriteTaskGroup( group *AWrok.TaskBatchGroup) (cid.Cid, error) {

	vfs.writeWaiter.Add(1)

	defer vfs.writeWaiter.Done()

	for k, v := range vfs.servies {
		if err := v.Shutdown(); err != nil {
			panic(err)
		}
		delete(vfs.servies, k)
	}

	var err error

	bmap := group.GetBatchMap()

	txs := &Transaction{
		transactions:make(map[string]*leveldb.Transaction),
	}

	var closeServices []AVdbComm.VDBSerices
	defer func() {

		for _, ser := range closeServices {

			if err := ser.Shutdown(); err != nil {

				log.Errorf("VDB Services closed error : %v", err)
			}
		}

	}()

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

		if services == nil {
			return cid.Undef, errors.New("DBKey nout found")
		}

		closeServices = append(closeServices, services)

		txs.transactions[dbkey], err = services.OpenTransaction()
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

	vfs.writeWaiter.Wait()

	vfs.writeWaiter.Add(1)
	defer vfs.writeWaiter.Done()

	for k, v := range vfs.servies {
		_ = v.Shutdown()
		delete(vfs.servies, k)
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
