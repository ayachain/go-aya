package block

import (
	"errors"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aBlocks struct {

	reader
	*mfs.Directory

	headAPI AIndexes.IndexesServices
	mfsstorage storage.Storage
	rawdb *leveldb.DB
}

func CreateServices( mdir *mfs.Directory, hapi AIndexes.IndexesServices, rdonly bool) Services {

	api := &aBlocks{
		Directory:mdir,
		headAPI:hapi,
	}

	api.rawdb, api.mfsstorage = AVdbComm.OpenExistedDB(mdir, DBPath, rdonly)

	return api
}

func (blks *aBlocks) GetBlocks( hashOrIndex...interface{} ) ([]*Block, error) {

	var blist []*Block

	for _, v := range hashOrIndex {

		var bhash EComm.Hash

		switch v.(type) {

		case uint64:

			hd, err := blks.headAPI.GetIndex(v.(uint64))
			if err != nil {
				return nil, err
			}
			bhash = hd.Hash

		case EComm.Hash:
			bhash = v.(EComm.Hash)

		default:
			return nil, errors.New("input params must be a index(uint64) or cid object")
		}

		dbval, err := blks.rawdb.Get( bhash.Bytes(), nil )
		if err != nil {
			return nil, err
		}

		subBlock := &Block{}
		if err := subBlock.Decode(dbval); err != nil {
			return nil, err
		}

		blist = append(blist, subBlock)
	}

	return blist, nil
}

func (blks *aBlocks) NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( blks.rawdb, blks.headAPI )
}

func (blks *aBlocks) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := blks.rawdb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (blks *aBlocks) Shutdown() error {

	if err := blks.rawdb.Close(); err != nil {
		return err
	}

	if err := blks.mfsstorage.Close(); err != nil {
		return err
	}

	return nil
}