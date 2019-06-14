package assets

import (
	"bytes"
	"encoding/binary"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
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

func (api *aAssetes) GetLockedTop100() ( []*SortAssets, error ) {

	list := make([]*SortAssets, 100)

	sbs := bytes.NewBuffer([]byte(DBTopIndexPrefix))
	sbs.Write( []byte{ 0x00, 0x00 } )

	ebs := bytes.NewBuffer([]byte(DBTopIndexPrefix))
	ebs.Write( []byte{ 0xff, 0xff })

	topIt := api.rawdb.NewIterator( &util.Range{Start:sbs.Bytes(), Limit:ebs.Bytes()}, nil )
	defer topIt.Release()

	for topIt.Next() {

		rcd, err := api.AssetsOf( EComm.BytesToAddress(topIt.Value()) )
		if err != nil {
			return nil, err
		}

		assets := &SortAssets{
			Address : EComm.BytesToAddress(topIt.Value()),
			Assets : rcd,
		}

		nbs := topIt.Key()[ len([]byte(DBTopIndexPrefix)) : ]

		index := int(binary.BigEndian.Uint16(nbs))

		if index < 0 || index >= 100 {
			return nil, errors.New("array index bound")
		}

		list[index] = assets
	}

	return list, nil
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