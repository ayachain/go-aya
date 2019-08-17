package block

import (
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aCache struct {

	writer

	cdb *leveldb.DB

	sourceReader *aBlocks

}

func newWriter( sread *aBlocks ) (MergeWriter, error) {

	memsto := storage.NewMemStorage()
	mdb, err := leveldb.Open(memsto, AvdbComm.OpenDBOpt)
	if err != nil {
		return nil, err
	}

	c := &aCache{
		sourceReader:sread,
		cdb:mdb,
	}

	return c, nil
}

/// Writer
func (cache *aCache) SetLatestPosBlockIndex( idx uint64 ) {
	_ = cache.cdb.Put(LatestPosBlockIdxKey, AvdbComm.BigEndianBytes(idx), AvdbComm.WriteOpt)
}

func (cache *aCache) AppendBlocks( blocks...*Block ) {

	if len(blocks) <= 0 {
		return
	}

	for _, v := range blocks {

		if err := cache.cdb.Put(v.GetHash().Bytes(), v.Encode(), AvdbComm.WriteOpt); err != nil {
			panic(err)
		}
	}

}

func (cache *aCache) WriteGenBlock( gen *GenBlock ) {

	hash := gen.GetHash().Bytes()

	if err := cache.cdb.Put(hash, gen.Encode(), AvdbComm.WriteOpt); err != nil {
		panic(err)
	}

}

func (cache *aCache) Close() {
	_ = cache.cdb.Close()
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

/// Reader mock impl
func (cache *aCache) GetLatestBlock() (*Block, error) {
	return cache.sourceReader.GetLatestBlock()
}

func (cache *aCache) GetBlocks( hashOrIndex...interface{} ) ([]*Block, error) {
	return cache.sourceReader.GetBlocks(hashOrIndex...)
}

func (cache *aCache) GetLatestPosBlockIndex( idx ... *indexes.Index ) uint64 {
	return cache.sourceReader.GetLatestPosBlockIndex()
}