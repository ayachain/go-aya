package receipt

import (
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
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


func (cache *aCache) HasTransactionReceipt( txhs EComm.Hash ) bool {

	st := append(txhs.Bytes(), AvdbComm.BigEndianBytes(0)... )
	ed := append(txhs.Bytes(), AvdbComm.BigEndianBytes((uint64(1) << 63) -1)... )

	s, err := cache.source.SizeOf([]util.Range{{Start:st,Limit:ed}})
	if err != nil {
		return false
	}

	return s.Sum() > 0
}

func (cache *aCache) GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error) {

	if !cache.HasTransactionReceipt(txhs) {
		return nil, leveldb.ErrNotFound
	}

	st := append(txhs.Bytes(), AvdbComm.BigEndianBytes(0)... )
	ed := append(txhs.Bytes(), AvdbComm.BigEndianBytes((uint64(1) << 63) -1)... )

	it := cache.source.NewIterator(&util.Range{Start:st,Limit:ed}, nil)

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

func (cache *aCache) Close() {
	_ = cache.cdb.Close()
}