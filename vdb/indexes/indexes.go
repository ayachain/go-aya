package indexes

import (
	"context"
	"github.com/ayachain/go-aya/vdb/common"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/pin"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	"github.com/ipfs/go-unixfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/whyrusleeping/go-logging"
	"sync"
)

var LatestIndexKey = []byte("LATEST")

var SyncWriteOpt = &opt.WriteOptions{Sync:true}

var log = logging.MustGetLogger("IndexesServices")

type aIndexes struct {

	IndexesServices
	AVdbComm.VDBSerices

	mfsstorage storage.Storage
	rawdb *leveldb.DB
	mfsroot *mfs.Root

	writeWaiter *sync.WaitGroup
}

func CreateServices( ind *core.IpfsNode, chainId string ) IndexesServices {

	adbpath := "/aya/chain/indexes/" + chainId

	var nd *merkledag.ProtoNode
	dsk := datastore.NewKey(adbpath)
	val, err := ind.Repo.Datastore().Get(dsk)

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

	root, err := mfs.NewRoot(
		context.TODO(),
		ind.DAG,
		nd,
		func(ctx context.Context, fcid cid.Cid) error {

			ind.Pinning.PinWithMode(fcid, pin.Any)

			return ind.Repo.Datastore().Put(dsk, fcid.Bytes())
		},
	)

	if err != nil {
		panic(err)
	}

	mfsdb, mfsstorage, err := AVdbComm.OpenDB( root.GetDirectory() )

	if err != nil {

		log.Error(err)

		if err := root.Close(); err != nil {
			panic(err)
		}

		root, err = mfs.NewRoot(
			context.TODO(),
			ind.DAG,
			unixfs.EmptyDirNode(),
			func(ctx context.Context, fcid cid.Cid) error {

				ind.Pinning.PinWithMode(fcid, pin.Any)

				return ind.Repo.Datastore().Put(dsk, fcid.Bytes())
			},
		)

		if err != nil {

			log.Error(err)
			panic(err)
		}

		mfsdb, mfsstorage, err = AVdbComm.OpenDB( root.GetDirectory() )

		if err != nil {

			log.Error(err)
			panic(err)
		}

	}

	api := &aIndexes{
		rawdb:mfsdb,
		mfsroot:root,
		mfsstorage:mfsstorage,
		writeWaiter:&sync.WaitGroup{},
	}

	return api
}

func ( i *aIndexes ) GetLatest() (*Index, error) {

	i.writeWaiter.Wait()

	bs, err := i.rawdb.Get([]byte(LatestIndexKey), nil)
	if err != nil {
		return nil, err
	}

	idx := &Index{}
	if err := idx.Decode(bs); err != nil {
		return nil, err
	}

	return idx, nil
}

func ( i *aIndexes ) GetIndex( blockNumber uint64 ) (*Index, error) {

	i.writeWaiter.Wait()

	key := common.BigEndianBytes(blockNumber)

	bs, err := i.rawdb.Get(key, nil)
	if err != nil {
		return nil, err
	}

	idx := &Index{}
	if err := idx.Decode(bs); err != nil {
		return nil, err
	}

	return idx,nil

}

func ( i *aIndexes ) Close() error {

	i.writeWaiter.Wait()

	i.writeWaiter.Add(1)
	defer i.writeWaiter.Done()
	if err := i.rawdb.Close(); err != nil {
		return err
	}

	if err := i.mfsstorage.Close(); err != nil {
		return err
	}

	if err := i.mfsroot.Close(); err != nil {
		return err
	}

	return nil
}

func ( i *aIndexes ) PutIndex( index *Index ) error {

	i.writeWaiter.Wait()

	key := common.BigEndianBytes(index.BlockIndex)
	value := index.Encode()

	if err := i.rawdb.Put( key, value, SyncWriteOpt); err != nil {
		return err
	}

	if err := i.rawdb.Put( []byte(LatestIndexKey), value, SyncWriteOpt ); err != nil {
		return err
	}

	return nil
}

func ( i *aIndexes ) PutIndexBy( num uint64, bhash EComm.Hash, ci cid.Cid ) error {

	//i.writeWaiter.Wait()

	key := common.BigEndianBytes(num)
	value := (&Index{
		BlockIndex:num,
		Hash:bhash,
		FullCID:ci,
	}).Encode()

	if err := i.rawdb.Put( key, value, SyncWriteOpt); err != nil {
		return err
	}

	if err := i.rawdb.Put( []byte(LatestIndexKey), value, SyncWriteOpt ); err != nil {
		return err
	}

	return nil
}

func ( i *aIndexes ) Flush() cid.Cid {

	i.writeWaiter.Wait()
	i.writeWaiter.Add(1)
	defer i.writeWaiter.Done()

	nd, err := mfs.FlushPath(context.TODO(), i.mfsroot, "/")

	if err != nil {
		return cid.Undef
	}

	return nd.Cid()
}