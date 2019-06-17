package transaction

import (
	"encoding/binary"
	"fmt"
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

func (cache *aCache) Put(tx *Transaction, bidx uint64) {

	countKey := append(TxCountPrefix, tx.From.Bytes()... )

	exist, err := AvdbComm.CacheHas(cache.source, cache.cdb, countKey)
	if err != nil {
		panic(err)
	}

	var txcount uint64 = 0
	if exist {

		cbs, err := AvdbComm.CacheGet(cache.source, cache.cdb, countKey)
		if err != nil {
			panic(err)
		}

		txcount = binary.BigEndian.Uint64(cbs)
	}

	key := append(tx.GetHash256().Bytes(), AvdbComm.BigEndianBytes(bidx)...)
	if err := cache.cdb.Put(key, tx.Encode(), nil); err != nil {
		panic(err)
	}

	txcount ++
	if err := cache.cdb.Put( countKey, AvdbComm.BigEndianBytes(txcount), nil ); err != nil {
		panic(err)
	}

}

func (cache *aCache) GetTxByHash( hash EComm.Hash ) (*Transaction, error) {

	tx := &Transaction{}

	val, err := AvdbComm.CacheGet(cache.source, cache.cdb, hash.Bytes())
	if err != nil {
		return nil, err
	}

	if err := tx.Decode(val); err != nil {
		return nil, fmt.Errorf("%v can't found transaction", hash.String())
	}

	return tx, nil
}

func (cache *aCache) GetTxByHashBs( hsbs []byte ) (*Transaction, error) {

	hash := EComm.BytesToHash(hsbs)

	return cache.GetTxByHash(hash)
}

func (cache *aCache) GetTxCount( address EComm.Address ) (uint64, error) {

	key := append(TxCountPrefix, address.Bytes()... )

	v, err := AvdbComm.CacheGet( cache.source, cache.cdb, key )

	if err != nil {

		if err == leveldb.ErrNotFound {
			return 0, nil
		} else {
			return 0, err
		}

	}

	return binary.BigEndian.Uint64(v), nil
}

func (cache *aCache) MergerBatch() *leveldb.Batch {

	batch := &leveldb.Batch{}

	it := cache.cdb.NewIterator(nil, nil)

	for it.Next() {

		batch.Put( it.Key(), it.Value() )

	}

	return batch
}
