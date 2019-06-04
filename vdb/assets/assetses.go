package assets

import (
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
)

type aAssetes struct {
	AssetsAPI
	*mfs.Root
	rawdb *leveldb.DB
}

func CreatePropertyAPI( rootref *mfs.Root ) AssetsAPI {

	api := &aAssetes{
		Root:rootref,
	}

	api.rawdb = common.OpenStandardDB(rootref, availDBPath)

	return api
}

func (api *aAssetes) AssetOf( key []byte ) ( *Assets, error ) {

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

	fromAsset, err := api.AssetOf(from)
	if err != nil {
		return nil,nil, err
	}

	toAsset, err := api.AssetOf(to)
	if err != nil {
		if err != os.ErrNotExist {
			return nil, nil, err
		}
	}

	if toAsset == nil {

		toAsset = &Assets{
			Version:DRVer,
			Avail:0,
			Vote:0,
			ExtraCid:cid.Undef,
		}

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