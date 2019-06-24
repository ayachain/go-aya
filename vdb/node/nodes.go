package node

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-mfs"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

type aNodes struct {

	reader
	*mfs.Directory

	mfsstorage storage.Storage
	ldb *leveldb.DB
	dbSnapshot *leveldb.Snapshot
	snLock sync.RWMutex
}


func CreateServices( mdir *mfs.Directory ) Services {

	var err error

	api := &aNodes{
		Directory:mdir,
	}

	api.ldb, api.mfsstorage = AVdbComm.OpenExistedDB(mdir, DBPath)

	api.dbSnapshot, err = api.ldb.GetSnapshot()
	if err != nil {
		_ = api.ldb.Close()
		log.Error(err)
		return nil
	}

	return api
}

func (api *aNodes) NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( api.dbSnapshot )
}

func (api *aNodes) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := api.ldb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}


func (api *aNodes) Shutdown() error {

	api.snLock.Lock()
	defer api.snLock.Unlock()

	if api.dbSnapshot != nil {
		api.dbSnapshot.Release()
	}

	if err := api.mfsstorage.Close(); err != nil {
		return err
	}

	if err := api.ldb.Close(); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (api *aNodes) Close() {
	_ = api.Shutdown()
}


func (api *aNodes) GetNodeByPeerId( peerId string ) (*Node, error) {

	api.snLock.RLock()
	defer api.snLock.RUnlock()

	bs, err := api.dbSnapshot.Get( []byte(peerId), nil )

	if err != nil {
		return nil, err
	}

	nd := &Node{}

	if err := nd.Decode(bs); err != nil {
		return nil, err
	}

	return nd, nil
}

func (api *aNodes) GetFirst() *Node {

	api.snLock.RLock()
	defer api.snLock.RUnlock()

	it := api.dbSnapshot.NewIterator( &util.Range{nil,nil}, nil )

	defer it.Release()

	var maxnd *Node

	for it.Next() {

		nd := &Node{}

		if err := nd.Decode(it.Value()); err != nil {
			continue
		}

		if maxnd == nil {

			maxnd = nd
			continue

		} else {

			if nd.Votes > maxnd.Votes {
				maxnd = nd
			}

		}

	}

	return maxnd
}

func (api *aNodes) GetLatest() *Node {

	api.snLock.RLock()
	defer api.snLock.RUnlock()

	it := api.dbSnapshot.NewIterator( &util.Range{nil,nil}, nil )

	defer it.Release()

	var minnd *Node

	for it.Next() {

		nd := &Node{}

		if err := nd.Decode(it.Value()); err != nil {
			continue
		}

		if minnd == nil {

			minnd = nd
			continue

		} else {

			if nd.Votes < minnd.Votes {
				minnd = nd
			}

		}

	}

	return minnd
}

func (api *aNodes) UpdateSnapshot() error {

	api.snLock.Lock()
	defer api.snLock.Unlock()

	if api.dbSnapshot != nil {
		api.dbSnapshot.Release()
	}

	var err error
	api.dbSnapshot, err = api.ldb.GetSnapshot()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}