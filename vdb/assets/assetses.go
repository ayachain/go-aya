package assets

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"sync"
)

type aAssetes struct {

	Services
	*mfs.Directory

	ldb *leveldb.DB

	dbSnapshot *leveldb.Snapshot
	snLock sync.RWMutex

	mfsstorage storage.Storage
}

func CreateServices( mdir *mfs.Directory ) Services {

	var err error
	api := &aAssetes{
		Directory:mdir,
	}

	api.ldb, api.mfsstorage = AVdbComm.OpenExistedDB( mdir, DBPATH )

	api.dbSnapshot, err = api.ldb.GetSnapshot()
	if err != nil {
		_ = api.ldb.Close()
		log.Error(err)
		return nil
	}

	return api
}

func (api *aAssetes) Shutdown() error {

	api.snLock.Lock()
	defer api.snLock.Unlock()

	if api.dbSnapshot != nil {
		api.dbSnapshot.Release()
	}

	if err := api.mfsstorage.Close(); err != nil {
		return err
	}

	if err := api.ldb.Close(); err != nil {
		return err
	}

	return nil
}


func (api *aAssetes) Close() {
	_ = api.Shutdown()
}


func (api *aAssetes) AssetsOf( addr EComm.Address ) ( *Assets, error ) {

	api.snLock.RLock()
	defer api.snLock.RUnlock()

	bnc, err := api.dbSnapshot.Get(addr.Bytes(), nil)

	if err != nil {
		return nil, err
	}

	rcd := &Assets{}
	if err := rcd.Decode(bnc); err != nil {
		return nil, err
	}

	return rcd, nil
}


func (api *aAssetes) NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( api.dbSnapshot )
}


func (api *aAssetes) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := api.ldb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}


func (api *aAssetes) UpdateSnapshot() error {

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