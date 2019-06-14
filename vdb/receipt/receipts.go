package receipt

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"sync"
)

type aReceipt struct {

	Services
	*mfs.Directory

	mfsstorage storage.Storage
	rawdb *leveldb.DB
	RWLocker sync.RWMutex
}

func CreateServices( mdir *mfs.Directory, rdonly bool ) Services {

	api := &aReceipt{
		Directory:mdir,
	}

	api.rawdb, api.mfsstorage = AVdbComm.OpenExistedDB(mdir, DBPath, rdonly)

	return api
}

func (r *aReceipt) GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error) {

	vbs, err := r.rawdb.Get( txhs.Bytes(), nil)

	if err != nil {
		return nil, err
	}

	rp := &Receipt{}

	err = rp.Decode( vbs )
	if err != nil {
		return nil, err
	}

	return rp, nil
}

func (r *aReceipt) Close() {

	_ = r.rawdb.Close()

	_ = r.mfsstorage.Close()

	_ = r.Flush()

}

func (r *aReceipt) NewCache() (AVdbComm.VDBCacheServices, error) {

	return newCache( r.rawdb )

}

func (r *aReceipt) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := r.rawdb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (r *aReceipt) Shutdown() error {

	_ = r.rawdb.Close()
	_ = r.mfsstorage.Close()

	return r.Flush()
}