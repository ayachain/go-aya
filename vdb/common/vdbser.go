package common

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
)

type VDBCacheServices interface {
	MergerBatch() *leveldb.Batch
	Close()
}


type VDBSerices interface {
	NewWriter() (VDBCacheServices, error)
}

func LookupDBPath( root *mfs.Root, path string ) (*mfs.Directory, error) {

	nd, err := mfs.Lookup(root, path)

	if err != nil {

		err := mfs.Mkdir(root, path, mfs.MkdirOpts{ Mkparents:true, Flush:true })
		if err != nil {
			return nil, err
		}

		nd, err = mfs.Lookup(root, path)
		if err != nil {
			return nil, err
		}

	}

	dir, ok := nd.(*mfs.Directory)
	if !ok {
		return nil, mfs.ErrInvalidChild
	}

	return dir, nil
}

func GetDBRoot( ctx context.Context, ind *core.IpfsNode, vcid cid.Cid, dbpath string ) ( dir *mfs.Directory, err error, closer func() ) {

	rnd, err := ind.DAG.Get(ctx, vcid)
	if err != nil {
		return nil, err, nil
	}

	pbnd, _ := rnd.(*merkledag.ProtoNode)
	mroot, err := mfs.NewRoot(ctx, ind.DAG, pbnd, func(i context.Context, i2 cid.Cid) error {
		return nil
	})
	if err != nil {

		if mroot != nil {
			_ = mroot.Close()
		}

		return nil, err, nil
	}

	vdir, err := LookupDBPath(mroot, dbpath)
	if err != nil {

		_ = mroot.Close()
		return nil, err, nil
	}

	return vdir, nil, func() {
		_ = mroot.Close()
	}

}