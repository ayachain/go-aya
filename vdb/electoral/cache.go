package electoral

import (
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aCache struct {

	writer

	source *leveldb.Snapshot
	cdb *leveldb.DB
}

func newCache( sourceDB *leveldb.Snapshot ) (Caches, error) {

	memsto := storage.NewMemStorage()

	mdb, err := leveldb.Open(memsto, AvdbComm.OpenDBOpt)
	if err != nil {
		return nil, err
	}

	c := &aCache{
		source:sourceDB,
		cdb:mdb,
	}

	return c, nil
}


func (cache *aCache) AppendPacker( packer *Electoral ) error {
	return cache.cdb.Put( AvdbComm.BigEndianBytes(packer.BlockIndex), packer.Address.Bytes(), AvdbComm.WriteOpt )
}


func (cache *aCache) PackerOf(index uint64) (*Electoral, error) {

	bs, err := AvdbComm.CacheGet( cache.source, cache.cdb, AvdbComm.BigEndianBytes(index) )

	if err != nil {
		return nil, err
	}

	ele := &Electoral{}

	if err = ele.Decode(bs); err != nil {
		return nil, err
	}

	return ele, nil
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