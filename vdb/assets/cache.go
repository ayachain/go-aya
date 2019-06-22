package assets

import (
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aCache struct {

	Caches

	source *leveldb.DB
	cdb *leveldb.DB
}


func newCache( sourceDB *leveldb.DB ) (Caches, error) {

	memsto := storage.NewMemStorage()

	mdb, err := leveldb.Open(memsto, nil)
	if err != nil {
		return nil, err
	}

	c := &aCache{
		source:sourceDB,
		cdb:mdb,
	}

	return c, nil
}

func (cache *aCache) AssetsOf( addr EComm.Address ) ( *Assets, error ) {

	bnc, err := AvdbComm.CacheGet( cache.source, cache.cdb, addr.Bytes() )
	if err != nil {
		return nil, err
	}

	rcd := &Assets{}
	if err := rcd.Decode(bnc); err != nil {
		return nil, err
	}

	return rcd, nil
}

func (cache *aCache) Close() {
	_ = cache.cdb.Close()
}

func (cache *aCache) MergerBatch() *leveldb.Batch {

	batch := &leveldb.Batch{}

	it := cache.cdb.NewIterator(nil, nil)

	for it.Next() {

		batch.Put( it.Key(), it.Value() )

	}

	return batch
}

func (cache *aCache) PutNewAssets( addr EComm.Address, ast *Assets ) {

	if err := cache.cdb.Put( addr.Bytes(), ast.Encode(), nil ); err != nil {
		panic(err)
	}
}

func (cache *aCache) Put( addr EComm.Address, ast *Assets ) {

	if err := cache.cdb.Put( addr.Bytes(), ast.Encode(), nil ); err != nil {
		panic(err)
	}
}