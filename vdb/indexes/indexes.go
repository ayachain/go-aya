package indexes

import (
	"context"
	ADB "github.com/ayachain/go-aya-alvm-adb"
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
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/whyrusleeping/go-logging"
	"sync"
)

var LatestIndexKey = []byte("LATEST")

var log = logging.MustGetLogger("IndexesServices")

/// Deve
const AIndexesKeyPathPrefix = "/aya/chain/indexes/dev/0807/4/"
/// Prod
//const AIndexesKeyPathPrefix = "/aya/chain/indexes/"

type aIndexes struct {

	IndexesServices
	AVdbComm.VDBSerices

	mfsroot *mfs.Root
	mfsstorage *ADB.MFSStorage

	ldb *leveldb.DB
	snLock sync.RWMutex

	ind *core.IpfsNode
	chainId string
}

func CreateServices( ind *core.IpfsNode, chainId string ) IndexesServices {

	adbpath := AIndexesKeyPathPrefix + chainId

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

	api := &aIndexes{
		ind:ind,
		chainId:chainId,
	}

	root, err := mfs.NewRoot(
		context.TODO(),
		ind.DAG,
		nd,
		func(ctx context.Context, fcid cid.Cid) error {
			ind.Pinning.PinWithMode(fcid, pin.Any)
			return nil
		},
	)

	if err != nil {
		panic(err)
	}

	mfsdb, mfsstorage, err := AVdbComm.OpenDB( root.GetDirectory(), "Indexes" )

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
				return nil
			},
		)

		if err != nil {

			log.Error(err)
			panic(err)
		}

		mfsdb, mfsstorage, err = AVdbComm.OpenDB( root.GetDirectory(), "Indexes" )

		if err != nil {

			log.Error(err)
			panic(err)
		}

	}

	api.ldb = mfsdb
	api.mfsroot = root
	api.mfsstorage = mfsstorage

	return api
}

func ( i *aIndexes ) GetLatest() (*Index, error) {

	i.snLock.RLock()
	defer i.snLock.RUnlock()

	bs, err := i.ldb.Get([]byte(LatestIndexKey), nil)
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

	i.snLock.RLock()
	defer i.snLock.RUnlock()

	key := common.BigEndianBytes(blockNumber)

	bs, err := i.ldb.Get(key, nil)
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

	i.snLock.Lock()
	defer i.snLock.Unlock()

	if err := i.ldb.Close(); err != nil {
		return err
	}

	if err := i.mfsroot.Close(); err != nil {
		return err
	}

	return nil
}

func ( i *aIndexes ) PutIndex( index *Index ) error {

	i.snLock.Lock()
	defer i.snLock.Unlock()

	key := common.BigEndianBytes(index.BlockIndex)

	value := index.Encode()

	dbtx, err := i.ldb.OpenTransaction()
	if err != nil {
		return err
	}

	if err := dbtx.Put(key, value, AVdbComm.WriteOpt); err != nil {
		return err
	}

	if err := dbtx.Put( []byte(LatestIndexKey), value, AVdbComm.WriteOpt ); err != nil {
		return err
	}

	if err := dbtx.Commit(); err !=nil {
		return err
	}

	return nil
}

func ( i *aIndexes ) PutIndexBy( num uint64, bhash EComm.Hash, ci cid.Cid ) error {

	i.snLock.Lock()
	defer i.snLock.Unlock()

	key := common.BigEndianBytes(num)
	value := (&Index{
		BlockIndex:num,
		Hash:bhash,
		FullCID:ci,
	}).Encode()

	dbtx, err := i.ldb.OpenTransaction()
	if err != nil {
		return err
	}

	if err := dbtx.Put( key, value, AVdbComm.WriteOpt); err != nil {
		return err
	}

	if err := dbtx.Put( []byte(LatestIndexKey), value, AVdbComm.WriteOpt ); err != nil {
		return err
	}

	if err := dbtx.Commit(); err !=nil {
		return err
	}

	return nil
}


func (api *aIndexes) UpdateSnapshot() error {
	return nil
}

func (api *aIndexes) Flush() cid.Cid {

	if err := api.ldb.CompactRange(util.Range{}); err != nil {
		log.Error(err)
	}

	nd, err := mfs.FlushPath( context.TODO(), api.mfsroot, "/")

	if err != nil {
		return cid.Undef
	}

	dsk := datastore.NewKey(AIndexesKeyPathPrefix + api.chainId)
	if err := api.ind.Repo.Datastore().Put(dsk, nd.Cid().Bytes()); err != nil {
		return cid.Undef
	} else {
		return nd.Cid()
	}
}