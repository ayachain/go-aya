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
	"github.com/whyrusleeping/go-logging"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

var log = logging.MustGetLogger("IndexesServices")

/// Dev
//const AIndexesKeyPathPrefix = "/aya/chain/indexes/dev/0810/15/"
/// Prod
const AIndexesKeyPathPrefix = "/aya/chain/indexes/"

const (
	idbFileNamePrefix    	= "page_"
	idbFileNameSuffix	  	= "idx"
	idbLatestIndex		  	= "latest.idx"
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

	log.Infof("Read Indexes DB : %v", nd.Cid().String())

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
	lin := binary.LittleEndian.Uint64(bs)

	return i.GetIndex(lin)
}

func ( i *aIndexes ) GetIndex( blockNumber uint64 ) (*Index, error) {

	i.snLock.RLock()
	defer i.snLock.RUnlock()

	page := blockNumber / 1024
	offset := blockNumber % 1024

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

	page := index.BlockIndex / 1024
	offset := index.BlockIndex % 1024

	fname := fmt.Sprintf("%v%d.%v", idbFileNamePrefix, page, idbFileNameSuffix)

	dir := i.mfsroot.GetDirectory()
	nd, err := dir.Child(fname)
	if err != nil {

		if err == os.ErrNotExist {

			//file not exist
			nnd := merkledag.NodeWithData(unixfs.FilePBData(nil, 1024 * StaticSize))
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

func ( i *aIndexes ) UpdateSnapshot() error {
	return nil
}

func ( i *aIndexes ) Flush() error {

	i.ind.Pinning.PinWithMode(i.latestCID, pin.Any)

	dsk := datastore.NewKey(AIndexesKeyPathPrefix + i.chainId)

	if err := i.ind.Repo.Datastore().Put( dsk, i.latestCID.Bytes() ); err != nil {
		return err
	} else {

		i.ind.Pinning.PinWithMode( i.latestCID, pin.Any )

		log.Infof("Save Indexes DB : %v", i.latestCID.String())
		return nil
	}

}

func ( i *aIndexes) SyncToCID( fullCID cid.Cid ) error {

	i.snLock.Lock()
	defer i.snLock.Unlock()

	i.latestCID = fullCID

	i.ind.Pinning.PinWithMode(i.latestCID, pin.Any)

	dsk := datastore.NewKey(AIndexesKeyPathPrefix + i.chainId)

	if err := i.ind.Repo.Datastore().Put( dsk, i.latestCID.Bytes() ); err != nil {
		return err
	} else {

		i.ind.Pinning.PinWithMode( i.latestCID, pin.Any )

		log.Infof("Sync Indexes DB : %v", i.latestCID.String())
		return nil
	}

}