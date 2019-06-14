package indexes

import (
	"context"
	"github.com/ayachain/go-aya/vdb/common"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	"github.com/ipfs/go-unixfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"time"
)

var LatestIndexKey = []byte("LATEST")

type aIndexes struct {

	IndexesServices

	*mfs.Directory
	AVdbComm.VDBSerices

	mfsstorage storage.Storage
	rawdb *leveldb.DB
	mfsroot *mfs.Root
}

func CreateServices( ind *core.IpfsNode, chainId string ) IndexesServices {

	adbpath := "/aya/chain/indexes/" + chainId + "t1"

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

		ctx, cancel := context.WithTimeout(context.Background(), time.Second  * 5 )
		defer cancel()

		rnd, err := ind.DAG.Get(ctx, c)
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
			return ind.Repo.Datastore().Put(dsk, fcid.Bytes())
		},
	)

	if err != nil {
		panic(err)
	}

	mfsdb, mfsstorage := AVdbComm.OpenExistedDB(root.GetDirectory(), "/", false)

	api := &aIndexes{
		rawdb:mfsdb,
		mfsroot:root,
		mfsstorage:mfsstorage,
	}

	return api
}

func ( i *aIndexes ) GetLatest() *Index {

	bs, err := i.rawdb.Get([]byte(LatestIndexKey), nil)
	if err != nil {
		return nil
	}

	idx := &Index{}
	if err := idx.Decode(bs); err != nil {
		return nil
	}

	return idx
}

func ( i *aIndexes ) GetIndex( blockNumber uint64 ) (*Index, error) {

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

	if err := i.rawdb.Close(); err != nil {
		return err
	}

	if err := i.mfsroot.Close(); err != nil {
		return err
	}

	return nil
}

func ( i *aIndexes ) PutIndex( index *Index ) error {

	key := common.BigEndianBytes(index.BlockIndex)
	value := index.Encode()

	if err := i.rawdb.Put(key, value, nil); err != nil {
		return err
	}

	if err := i.rawdb.Put( []byte(LatestIndexKey), value, nil ); err != nil {
		return err
	}

	return nil
}

func ( i *aIndexes ) PutIndexBy( num uint64, bhash EComm.Hash, ci cid.Cid ) error {

	key := common.BigEndianBytes(num)
	value := (&Index{
		BlockIndex:num,
		Hash:bhash,
		FullCID:ci,
	}).Encode()

	if err := i.rawdb.Put(key, value, nil); err != nil {
		return err
	}

	if err := i.rawdb.Put( []byte(LatestIndexKey), value, nil ); err != nil {
		return err
	}

	return nil
}