package assets

import (
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"sync"
)

type aCache struct {

	MergeWriter

	sourceReader *aAssetes

	cdb *leveldb.DB

	snLock sync.RWMutex
}

func newWriter( sread *aAssetes ) (MergeWriter, error) {

	memsto := storage.NewMemStorage()

	mdb, err := leveldb.Open(memsto, nil)

	if err != nil {
		return nil, err
	}

	c := &aCache{
		sourceReader:sread,
		cdb:mdb,
	}

	return c, nil
}

func (c *aCache) Close() {

	c.snLock.Lock()
	defer c.snLock.Unlock()

	_ = c.cdb.Close()
}

func (c *aCache) MergerBatch() *leveldb.Batch {

	c.snLock.Lock()
	defer c.snLock.Unlock()

	batch := &leveldb.Batch{}

	it := c.cdb.NewIterator(nil, nil)
	defer it.Release()

	for it.Next() {
		batch.Put( it.Key(), it.Value() )
	}

	return batch
}

func (c *aCache) Put( addr EComm.Address, ast *Assets ) {

	c.snLock.Lock()
	defer c.snLock.Unlock()

	if err := c.cdb.Put( addr.Bytes(), ast.Encode(), AvdbComm.WriteOpt ); err != nil {
		panic(err)
	}

}

func (c *aCache) AssetsOf( addr EComm.Address, idx ... *indexes.Index ) ( *Assets, error ) {

	c.snLock.RLock()
	defer c.snLock.RUnlock()

	inCache, err := c.cdb.Has(addr.Bytes(), nil)
	if err != nil && err != leveldb.ErrNotFound{
		return nil, err
	}

	if !inCache {

		return c.sourceReader.AssetsOf(addr)

	} else {

		bnc, err := c.cdb.Get(addr.Bytes(), nil)
		if err != nil {
			panic(err)
		}

		rcd := &Assets{}
		if err := rcd.Decode(bnc); err != nil {
			return nil, err
		}

		return rcd, nil
	}
}