package receipt

import (
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aCache struct {

	MergeWriter

	sourceReader *aReceipt

	cdb *leveldb.DB
}


func newWriter( sreader *aReceipt ) (MergeWriter, error) {

	memsto := storage.NewMemStorage()

	mdb, err := leveldb.Open(memsto, AvdbComm.OpenDBOpt)
	if err != nil {
		return nil, err
	}

	c := &aCache{
		sourceReader:sreader,
		cdb:mdb,
	}

	return c, nil
}

func (cache *aCache) MergerBatch() *leveldb.Batch {

	batch := &leveldb.Batch{}

	it := cache.cdb.NewIterator(nil, nil)
	defer it.Release()

	for it.Next() {
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

/// Reader mock impl
func (cache *aCache) GetTransactionReceipt( txhs EComm.Hash, idx ... *indexes.Index ) (*Receipt, error) {
	return cache.sourceReader.GetTransactionReceipt(txhs)
}

func (cache *aCache) HasTransactionReceipt( txhs EComm.Hash, idx ... *indexes.Index ) bool {
	return cache.sourceReader.HasTransactionReceipt( txhs )
}