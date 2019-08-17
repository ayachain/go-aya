package assets

import (
	"context"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-ipfs/core"
	"github.com/syndtr/goleveldb/leveldb"
)

type aAssetes struct {

	Services

	idxs indexes.IndexesServices

	ind *core.IpfsNode
}

func CreateServices( ind *core.IpfsNode, idxServices indexes.IndexesServices ) Services {
	return &aAssetes{
		idxs:idxServices,
		ind:ind,
	}
}

func (api *aAssetes) AssetsOf( addr EComm.Address, idx... *indexes.Index ) ( *Assets, error ) {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = api.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), api.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	rcd := &Assets{}
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB ) error {

		bnc, err := db.Get(addr.Bytes(), nil)
		if err != nil {
			return  err
		}

		if err := rcd.Decode(bnc); err != nil {
			return err
		}

		return nil

	}, DBPath); err != nil {
		return nil, err
	}

	return rcd, nil
}


func (api *aAssetes) NewWriter() (AVdbComm.VDBCacheServices, error) {
	return newWriter( api )
}