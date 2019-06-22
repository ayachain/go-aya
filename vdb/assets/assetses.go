package assets

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aAssetes struct {

	Services

	*mfs.Directory
	rawdb *leveldb.DB
	mfsstorage storage.Storage
}

func CreateServices( mdir *mfs.Directory, rdOnly bool ) Services {

	api := &aAssetes{
		Directory:mdir,
	}

	api.rawdb, api.mfsstorage = AVdbComm.OpenExistedDB( mdir, DBPATH, rdOnly )

	return api
}

func (api *aAssetes) Shutdown() error {

	_ = api.rawdb.Close()
	_ = api.mfsstorage.Close()

	return api.Flush()
}

func (api *aAssetes) Close() {

	_ = api.rawdb.Close()
	_ = api.mfsstorage.Close()

}

func (api *aAssetes) AssetsOf( addr EComm.Address ) ( *Assets, error ) {

	bnc, err := api.rawdb.Get(addr.Bytes(), nil)
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
	return newCache( api.rawdb )
}

func (api *aAssetes) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := api.rawdb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}