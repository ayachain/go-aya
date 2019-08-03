package block

import (
	"encoding/binary"
	"errors"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aCache struct {

	writer

	headAPI AIndexes.IndexesServices
	source *leveldb.Snapshot
	cdb *leveldb.DB
}

func newCache( sourceDB *leveldb.Snapshot, idxReader AIndexes.IndexesServices ) (Caches, error) {

	memsto := storage.NewMemStorage()

	mdb, err := leveldb.Open(memsto, AvdbComm.OpenDBOpt)
	if err != nil {
		return nil, err
	}

	c := &aCache{
		source:sourceDB,
		cdb:mdb,
		headAPI:idxReader,
	}

	return c, nil
}

func (cache *aCache) GetLatestPosBlockIndex() uint64 {

	if exist, err := cache.source.Has(LatestPosBlockIdxKey, nil); err != nil {

		panic(err)

	} else if !exist {

		return 0

	} else {

		if bs, err := cache.source.Get(LatestPosBlockIdxKey, nil); err != nil {

			return 0

		} else {

			return binary.BigEndian.Uint64(bs)

		}

	}

}

func (cache *aCache) SetLatestPosBlockIndex( idx uint64 ) {
	_ = cache.cdb.Put(LatestPosBlockIdxKey, AvdbComm.BigEndianBytes(idx), AvdbComm.WriteOpt)
}

func (cache *aCache) GetBlocks( hashOrIndex...interface{} ) ([]*Block, error) {

	var blist []*Block

	for _, v := range hashOrIndex {

		var bhash EComm.Hash

		switch v.(type) {

		case uint64:

			hd, err := cache.headAPI.GetIndex(v.(uint64))
			if err != nil {
				return nil, err
			}
			bhash = hd.Hash


		case EComm.Hash:
			bhash = v.(EComm.Hash)


		default:
			return nil, errors.New("input params must be a index(uint64) or cid object")
		}

		dbval, err := AvdbComm.CacheGet( cache.source, cache.cdb, bhash.Bytes() )
		if err != nil {
			return nil, err
		}

		subBlock := &Block{}
		if err := subBlock.Decode(dbval); err != nil {
			return nil, err
		}

		blist = append(blist, subBlock)
	}

	return blist, nil
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

	for it.Next() {

		log.Infof("BatchPut ValueHash:%v)", crypto.Keccak256Hash(it.Value()).String() )

		batch.Put( it.Key(), it.Value() )

	}

	return batch
}