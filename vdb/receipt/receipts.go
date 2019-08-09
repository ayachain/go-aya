package receipt

import (
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

type aReceipt struct {

	Services
	*mfs.Directory

	mfsstorage *ADB.MFSStorage
	ldb *leveldb.DB
	dbSnapshot *leveldb.Snapshot
	snLock sync.RWMutex
}

func CreateServices( mdir *mfs.Directory ) Services {

	var err error

	api := &aReceipt{
		Directory:mdir,
	}

	api.ldb, api.mfsstorage, err = AVdbComm.OpenExistedDB(mdir, DBPath)
	if err != nil {
		panic(err)
	}

	api.dbSnapshot, err = api.ldb.GetSnapshot()
	if err != nil {
		log.Error(err)
		return nil
	}

	return api
}

func (r *aReceipt) HasTransactionReceipt( txhs EComm.Hash ) bool {

	it := r.ldb.NewIterator( util.BytesPrefix(txhs.Bytes()), nil )

	defer it.Release()

	if it.Next() {
		return true
	}

	return false
}

func (r *aReceipt) GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error) {

	r.snLock.RLock()
	defer r.snLock.RUnlock()

	if !r.HasTransactionReceipt(txhs) {
		return nil, leveldb.ErrNotFound
	}

	it := r.ldb.NewIterator( util.BytesPrefix(txhs.Bytes()), nil )

	if !it.Next() {
		return nil, leveldb.ErrNotFound
	}

	rp := &Receipt{}

	err := rp.Decode( it.Value() )
	if err != nil {
		return nil, err
	}

	return rp, nil
}

func (r *aReceipt) Close() {
	_ = r.Shutdown()
}

func (r *aReceipt) NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( r.dbSnapshot )
}

func (r *aReceipt) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := r.ldb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (r *aReceipt) Shutdown() error {

	r.snLock.Lock()
	defer r.snLock.Unlock()

	if r.dbSnapshot != nil {
		r.dbSnapshot.Release()
	}

	//if err := r.mfsstorage.Close(); err != nil {
	//	return err
	//}

	if err := r.ldb.Close(); err != nil {
		log.Error(err)
		return err
	}

	return nil
}


func (api *aReceipt) UpdateSnapshot() error {

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

func (api *aReceipt) SyncCache() error {

	if err := api.ldb.CompactRange(util.Range{nil,nil}); err != nil {
		log.Error(err)
	}

	return api.mfsstorage.Flush()
}