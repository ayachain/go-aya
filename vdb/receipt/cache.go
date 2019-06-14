package receipt

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



func (cache *aCache) GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error) {

	vbs, err := AvdbComm.CacheGet(cache.source, cache.cdb, txhs.Bytes())

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

func (cache *aCache) MergerBatch() *leveldb.Batch {

	batch := &leveldb.Batch{}

	it := cache.cdb.NewIterator(nil, nil)

	for it.Next() {

		batch.Put( it.Key(), it.Value() )

	}

	return batch
}

func (cache *aCache) Put( txhs EComm.Hash, bindex uint64, receipt []byte ) {

	key := append(txhs.Bytes(), AvdbComm.BigEndianBytes(bindex)... )

	if err := cache.cdb.Put( key, receipt, nil ); err != nil {
		panic(err)
	}

}