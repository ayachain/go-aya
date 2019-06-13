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
	ReceiptsAPI
	*mfs.Directory

	mfsstorage storage.Storage
	rawdb *leveldb.DB
	RWLocker sync.RWMutex
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

func (txs *aReceipt) DBKey()	string {
	return DBPath
}

func CreateServices( mdir *mfs.Directory, rdonly bool ) ReceiptsAPI {

	api := &aReceipt{
		Directory:mdir,
	}

	api.rawdb, api.mfsstorage = AVdbComm.OpenExistedDB(mdir, DBPath, rdonly)

	return api
}

func (api *aReceipt) OpenVDBTransaction() (*leveldb.Transaction, *sync.RWMutex, error) {

	tx, err := api.rawdb.OpenTransaction()
	if err != nil {
		return nil, nil, err
	}

	return tx, &api.RWLocker, nil
}

func (api *aReceipt) Close() {

	api.RWLocker.Lock()
	defer api.RWLocker.Unlock()

	_ = api.rawdb.Close()
	_ = api.mfsstorage.Close()
	_ = api.Flush()

}