package assets

import (
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"sync"
)

type aAssetes struct {
	AssetsAPI
	*mfs.Directory

	RWLocker sync.RWMutex

	rawdb *leveldb.DB
}

func CreateServices( mdir *mfs.Directory ) AssetsAPI {

	api := &aAssetes{
		Directory:mdir,
	}

	api.rawdb = common.OpenExistedDB(mdir, DBPATH)

	return api
}

func (api *aAssetes) DBKey() string {
	return DBPATH
}

func (api *aAssetes) AssetsOf( key []byte ) ( *Assets, error ) {

	bnc, err := api.rawdb.Get(key, nil)
	if err != nil {
		return nil, err
	}

	rcd := &Assets{}
	if err := rcd.Decode(bnc); err != nil {
		return nil, err
	}

	return rcd, nil
}

func (api *aAssetes) AvailBalanceMove( from, to []byte, v uint64 ) ( aftf, aftt *Assets, err error ) {

	fromAsset, err := api.AssetsOf(from)
	if err != nil {
		return nil,nil, err
	}

	toAsset, err := api.AssetsOf(to)
	if err != nil {
		if err != os.ErrNotExist {
			return nil, nil, err
		}
	}

	if toAsset == nil {
		toAsset = NewAssets(0,0,0)
	}

	if fromAsset.Avail - fromAsset.Vote < v {
		return nil, nil, notEnoughError
	}

	fromAsset.Avail -= v
	fromAsset.Vote -= v

	toAsset.Avail += v
	toAsset.Vote += v

	mvBatch := &leveldb.Batch{}
	mvBatch.Put( from, fromAsset.Encode() )
	mvBatch.Put( to, fromAsset.Encode() )

	if err := api.rawdb.Write(mvBatch, nil); err != nil {
		return nil,nil, err
	}

	return fromAsset, toAsset, nil
}

func (api *aAssetes) OpenVDBTransaction() (*leveldb.Transaction, *sync.RWMutex, error) {

	tx, err := api.rawdb.OpenTransaction()
	if err != nil {
		return nil, nil, err
	}

	return tx, &api.RWLocker, nil
}


func (api *aAssetes) Close() {

	api.RWLocker.Lock()
	defer api.RWLocker.Unlock()

	_ = api.rawdb.Close()

}