package transaction

import (
	"encoding/binary"
	"fmt"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type aCache struct {

	Caches

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

func (cache *aCache) Put(tx *Transaction, bidx uint64) {

	countKey := append(TxOutCountPrefix, tx.From.Bytes()... )

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

	tx.BlockIndex = bidx

	if err := cache.cdb.Put(key, tx.Encode(), AvdbComm.WriteOpt); err != nil {
		panic(err)
	}

	// Block index resolve
	tx.BlockIndex = 0

	txcount ++

	if err := cache.cdb.Put( countKey, AvdbComm.BigEndianBytes(txcount), AvdbComm.WriteOpt ); err != nil {
		panic(err)
	}

	// Write from tx history
	fromTxTotalKey := append(TxTotalCountPrefix, tx.From.Bytes()...)
	toTxTotalKey := append(TxTotalCountPrefix, tx.To.Bytes()...)

	var (
		fromTxTotal uint64 = 0
		toTxTotal uint64 = 0
	)

	// from tx total key exist
	fttke, err := AvdbComm.CacheHas(cache.source, cache.cdb, fromTxTotalKey)
	if err != nil {

		panic(err)

	} else {

		if fttke {

			cbs, err := AvdbComm.CacheGet(cache.source, cache.cdb, fromTxTotalKey)
			if err != nil {
				panic(err)
			}

			fromTxTotal = binary.BigEndian.Uint64(cbs)

		}

		hkey := append(TxHistoryPrefix, tx.From.Bytes()... )
		hkey = append(hkey, AvdbComm.BigEndianBytes(fromTxTotal)...)

		fromTxTotal ++

		// Update total count
		if err := cache.cdb.Put( fromTxTotalKey, AvdbComm.BigEndianBytes(fromTxTotal), AvdbComm.WriteOpt ); err != nil {
			panic(err)
		}

		// insert history indexes
		if err := cache.cdb.Put( hkey, key, AvdbComm.WriteOpt ); err != nil {
			panic(err)
		}

	}

	// to tx total key exist
	tttke, err :=  AvdbComm.CacheHas(cache.source, cache.cdb, toTxTotalKey)
	if err != nil {

		panic(err)

	} else {

		if tttke {

			cbs, err := AvdbComm.CacheGet(cache.source, cache.cdb, toTxTotalKey)
			if err != nil {
				panic(err)
			}

			toTxTotal = binary.BigEndian.Uint64(cbs)
		}

		hkey := append(TxHistoryPrefix, tx.To.Bytes()... )
		hkey = append(hkey, AvdbComm.BigEndianBytes(toTxTotal)...)

		toTxTotal ++

		// Update total count
		if err := cache.cdb.Put( toTxTotalKey, AvdbComm.BigEndianBytes(toTxTotal), AvdbComm.WriteOpt ); err != nil {
			panic(err)
		}

		// insert history indexes
		if err := cache.cdb.Put( hkey, key, AvdbComm.WriteOpt ); err != nil {
			panic(err)
		}

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

	key := append(TxOutCountPrefix, address.Bytes()... )

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


func (cache *aCache) GetHistoryHash( address EComm.Address, offset uint64, size uint64) []EComm.Hash {

	itkey := append(TxHistoryPrefix, address.Bytes()...)

	it := cache.source.NewIterator( util.BytesPrefix(itkey), nil )

	var hashs []EComm.Hash

	for it.Next() {
		hashs = append(hashs, EComm.BytesToHash(it.Value()))
	}

	return hashs

}

func (cache *aCache) GetHistoryContent( address EComm.Address, offset uint64, size uint64) ([]*Transaction, error) {

	itkey := append(TxHistoryPrefix, address.Bytes()...)

	it := cache.source.NewIterator( util.BytesPrefix(itkey), nil )
	defer it.Release()

	var txs []*Transaction

	if offset > 0 {

		seekKey := append(itkey, AvdbComm.BigEndianBytes(offset)...)

		if !it.Seek(seekKey) {
			return nil, errors.New("seek history key expected ")
		}
	}

	s := uint64(0)

	for it.Next() {

		if s >= size - 1 {
			break
		}

		bs, err := cache.source.Get(it.Value(), nil)
		if err != nil {
			return nil, err
		}

		tx := &Transaction{}

		if err := tx.Decode(bs); err != nil {
			return nil, err
		}

		txs = append(txs, tx)

		s ++
	}

	return txs, nil

}