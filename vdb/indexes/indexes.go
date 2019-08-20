package indexes

import (
	"context"
	"encoding/binary"
	"fmt"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/pin"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	"github.com/ipfs/go-unixfs"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

const AIndexesKeyPathPrefix = "/aya/chain/indexes/"

const (
	idbFileNamePrefix    	= "page_"
	idbFileNameSuffix	  	= "idx"
	idbLatestIndex		  	= "latest.idx"
	idxPageIndexSize		= 5120
)

var (
	ErrCreateFailed	= errors.New("create index services failed")
	ErrSyncRollback = errors.New("sync to target cid has rollback")
)

type aIndexes struct {

	IndexesServices
	AVdbComm.VDBSerices

	mfsroot *mfs.Root
	snLock sync.RWMutex

	ind *core.IpfsNode
	chainId string

	latestCID cid.Cid
}

func CreateServices( ind *core.IpfsNode, chainId string, rcp bool ) IndexesServices {

	adbpath := AIndexesKeyPathPrefix + chainId

	var nd *merkledag.ProtoNode
	dsk := datastore.NewKey(adbpath)

	if !rcp {

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

	} else {

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
		func(ctx context.Context, cid cid.Cid) error {
			return nil
		},
	)

	if err != nil {
		log.Error(err)
		return nil
	}

	api.mfsroot = root

	return api
}

func ( i *aIndexes ) GetLatest() (*Index, error) {

	i.snLock.RLock()
	defer i.snLock.RUnlock()

	nd, err := i.mfsroot.GetDirectory().Child(idbLatestIndex)
	if err != nil {
		if err == os.ErrNotExist {
			return nil, nil
		} else {
			return nil, err
		}
	}

	fi, ok := nd.(*mfs.File)
	if !ok {
		return nil, fmt.Errorf("target /%v is not a file", idbLatestIndex)
	}

	fd, err := fi.Open(mfs.Flags{Read:true,Sync:false})
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	bs, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	// latest block index number by int64
	lin := binary.BigEndian.Uint64(bs)

	return i.GetIndex(lin)
}

func ( i *aIndexes ) GetIndex( blockNumber uint64 ) (*Index, error) {

	if blockNumber == ^uint64(0) {
		return i.GetLatest()
	}

	i.snLock.RLock()
	defer i.snLock.RUnlock()

	page := blockNumber / idxPageIndexSize
	offset := blockNumber % idxPageIndexSize

	fname := fmt.Sprintf("%v%d.%v", idbFileNamePrefix, page, idbFileNameSuffix)

	nd, err := i.mfsroot.GetDirectory().Child(fname)
	if err != nil {
		return nil, err
	}

	fi, ok := nd.(*mfs.File)
	if !ok {
		return nil, fmt.Errorf("target /%v is not a file", idbLatestIndex)
	}

	fd, err := fi.Open(mfs.Flags{Read:true,Sync:false})
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	if _, err := fd.Seek( int64(offset) * StaticSize,io.SeekStart); err != nil {
		return nil, err
	}

	idxbs := make([]byte, StaticSize)
	if _, err := fd.Read(idxbs); err != nil {
		return nil, err
	}

	idx := &Index{}
	if err := idx.Decode(idxbs); err != nil {
		return nil, err
	}

	return idx, nil
}

func ( i *aIndexes ) Close() error {

	i.snLock.Lock()
	defer i.snLock.Unlock()

	return i.mfsroot.Close()
}

func ( i *aIndexes ) PutIndex( index *Index ) (cid.Cid, error) {

	i.snLock.Lock()
	defer i.snLock.Unlock()

	page := index.BlockIndex / idxPageIndexSize
	offset := index.BlockIndex % idxPageIndexSize

	fname := fmt.Sprintf("%v%d.%v", idbFileNamePrefix, page, idbFileNameSuffix)

	dir := i.mfsroot.GetDirectory()
	nd, err := dir.Child(fname)
	if err != nil {

		if err == os.ErrNotExist {

			//file not exist
			nnd := merkledag.NodeWithData(unixfs.FilePBData(nil, idxPageIndexSize * StaticSize))
			nnd.SetCidBuilder(dir.GetCidBuilder())

			if err := dir.AddChild( fname, nnd ); err != nil {
				return cid.Undef, err
			}

			nd, err = dir.Child(fname)
			if err != nil {
				return cid.Undef, err
			}

		} else {

			return cid.Undef, err

		}
	}

	fi, ok := nd.(*mfs.File)
	if !ok {
		return cid.Undef, fmt.Errorf("target /%v is not a file", idbLatestIndex)
	}

	fd, err := fi.Open(mfs.Flags{Write:true, Sync:true})
	if err != nil {
		return cid.Undef, err
	}

	value := index.Encode()
	if _, err := fd.WriteAt(value, int64(offset) * StaticSize); err != nil {
		return cid.Undef, err
	}

	if err := fd.Close(); err != nil {
		log.Errorf("close idx error: %v", err)
		return cid.Undef, err
	}

	/// write latest index number
	if err := i.putLatestIndex(index.BlockIndex); err != nil {
		return cid.Undef, err
	}

	cnd, err := mfs.FlushPath(context.TODO(), i.mfsroot, "/")
	if err != nil {
		return cid.Undef, err
	}

	i.latestCID = cnd.Cid()

	return i.latestCID, nil
}

func ( i *aIndexes ) putLatestIndex( num uint64 ) error {

	dir := i.mfsroot.GetDirectory()
	nd, err := dir.Child(idbLatestIndex)
	if err != nil {

		if err == os.ErrNotExist {

			//file not exist
			nnd := merkledag.NodeWithData(unixfs.FilePBData(nil, 8))
			nnd.SetCidBuilder(dir.GetCidBuilder())

			if err := dir.AddChild( idbLatestIndex, nnd ); err != nil {
				return err
			}

			nd, err = dir.Child(idbLatestIndex)
			if err != nil {
				return err
			}

		}

	}

	fi, ok := nd.(*mfs.File)
	if !ok {
		return fmt.Errorf("target /%v is not a file", idbLatestIndex)
	}

	fd, err := fi.Open(mfs.Flags{Write:true,Sync:true})
	if err != nil {
		return err
	}
	defer fd.Close()

	if err := fd.Truncate(0); err != nil {
		return err
	}

	if _, err := fd.WriteAt( AVdbComm.BigEndianBytes(num), 0 ); err != nil {
		return err
	}

	return nil
}

func ( i *aIndexes ) PutIndexBy( num uint64, bhash EComm.Hash, ci cid.Cid ) (cid.Cid, error) {

	return i.PutIndex( &Index{
		BlockIndex:num,
		Hash:bhash,
		FullCID:ci,
	})

}

func ( i *aIndexes ) SyncToCID( fullCID cid.Cid ) error {

	i.snLock.Lock()
	defer i.snLock.Unlock()

	// rollback cid and mfs root
	rbcid := i.latestCID
	rbroot := i.mfsroot

	dsk := datastore.NewKey(AIndexesKeyPathPrefix + i.chainId)
	rnd, err := i.ind.DAG.Get(context.TODO(), fullCID)
	if err != nil {
		return err
	}

	pbnd, ok := rnd.(*merkledag.ProtoNode)
	if !ok {
		return errors.New("target cid is not a proto node")
	}

	newRoot, err := mfs.NewRoot(
		context.TODO(),
		i.ind.DAG,
		pbnd,
		func(ctx context.Context, cid cid.Cid) error {
			return nil
		},
	)
	if err != nil {
		return err
	}

	// try change date if has error, must jump to 'NeedRollBack' tag
	i.mfsroot = newRoot
	i.latestCID = fullCID

	// test read latest index
	lidx, success := i.simpleVerifyNoLock()
	if !success {
		goto NeedRollBack
	}

	i.ind.Pinning.PinWithMode(i.latestCID, pin.Any)

	if err := i.ind.Repo.Datastore().Put( dsk, i.latestCID.Bytes() ); err != nil {
		goto NeedRollBack
	}

	log.Infof("SyncTo: %08d - %v", lidx.BlockIndex, i.latestCID.String())

	_ = rbroot.Close()

	return nil

NeedRollBack:

	_ = i.mfsroot.Close()

	i.mfsroot = rbroot
	i.latestCID = rbcid

	return ErrSyncRollback
}

func ( i *aIndexes ) simpleVerifyNoLock() (*Index, bool) {

	nd, err := i.mfsroot.GetDirectory().Child(idbLatestIndex)
	if err != nil {
		if err == os.ErrNotExist {
			return nil, false
		} else {
			return nil, false
		}
	}

	fi, ok := nd.(*mfs.File)
	if !ok {
		return nil, false
	}

	fd, err := fi.Open(mfs.Flags{Read:true,Sync:false})
	if err != nil {
		return nil, false
	}
	defer fd.Close()

	bs, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, false
	}

	// latest block index number by int64
	lin := binary.BigEndian.Uint64(bs)

	// read index
	page := lin / idxPageIndexSize
	offset := lin % idxPageIndexSize

	fname := fmt.Sprintf("%v%d.%v", idbFileNamePrefix, page, idbFileNameSuffix)

	nd, err = i.mfsroot.GetDirectory().Child(fname)
	if err != nil {
		return nil, false
	}

	fi, ok = nd.(*mfs.File)
	if !ok {
		return nil, false
	}

	fd, err = fi.Open(mfs.Flags{Read:true,Sync:false})
	if err != nil {
		return nil, false
	}
	defer fd.Close()

	if _, err := fd.Seek( int64(offset) * StaticSize,io.SeekStart); err != nil {
		return nil, false
	}

	idxbs := make([]byte, StaticSize)
	if _, err := fd.Read(idxbs); err != nil {
		return nil, false
	}

	idx := &Index{}
	if err := idx.Decode(idxbs); err != nil {
		return nil, false
	}

	return idx, true
}

func ForkMerge( ind *core.IpfsNode, chainId string, index *Index ) (cid.Cid, error) {

	forkIdx := CreateServices(ind, chainId, false)
	if forkIdx == nil {
		return cid.Undef, ErrCreateFailed
	}
	defer func() {
		_ = forkIdx.Close()
	}()

	return forkIdx.PutIndex(index)
}