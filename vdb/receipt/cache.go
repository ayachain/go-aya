package receipt

import (
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prometheus/common/log"
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


func (cache *aCache) HasTransactionReceipt( txhs EComm.Hash ) bool {

	it := cache.source.NewIterator( util.BytesPrefix(txhs.Bytes()), nil )

	defer it.Release()

	if it.Next() {
		return true
	}

	return false
}


func (cache *aCache) GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error) {

	if !cache.HasTransactionReceipt(txhs) {
		return nil, leveldb.ErrNotFound
	}

	it := cache.source.NewIterator( util.BytesPrefix(txhs.Bytes()), nil )

	if !it.Next() {
		return nil, leveldb.ErrNotFound
	}

	rp := &Receipt{}

	err := rp.Decode( it.Value() )
	if err != nil {
		return nil, err
	}

	return rp, nil
}


func (cache *aCache) MergerBatch() *leveldb.Batch {

	batch := &leveldb.Batch{}

	it := cache.cdb.NewIterator(nil, nil)

	for it.Next() {

		log.Infof("BatchPut ValueHash:%v)", crypto.Keccak256Hash(it.Value()).String() )

		batch.Put( it.Key(), it.Value() )

	}

	return batch
}


func (cache *aCache) Put( txhs EComm.Hash, bindex uint64, receipt []byte ) {

	key := append(txhs.Bytes(), AvdbComm.BigEndianBytes(bindex)... )

	if err := cache.cdb.Put( key, receipt, AvdbComm.WriteOpt ); err != nil {
		panic(err)
	}

}


func (cache *aCache) Close() {
	_ = cache.cdb.Close()
}