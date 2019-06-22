package node

import (
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type aCache struct {

	writer

	headAPI AIndexes.IndexesServices
	source *leveldb.DB
	cdb *leveldb.DB

	delKeys [][]byte
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

func (cache *aCache) GetNodeByPeerId( peerId string ) (*Node, error) {

	bs, err := AvdbComm.CacheGet( cache.source, cache.cdb, []byte(peerId) )
	if err != nil {
		return nil, err
	}

	nd := &Node{}

	if err := nd.Decode(bs); err != nil {
		return nil, err
	}

	return nd, nil
}


func (cache *aCache) Close() {
	_ = cache.cdb.Close()
}


func (cache *aCache) MergerBatch() *leveldb.Batch {

	batch := &leveldb.Batch{}

	it := cache.cdb.NewIterator(nil, nil)

	for _, delk := range cache.delKeys {
		batch.Delete(delk)
	}

	for it.Next() {

		batch.Put( it.Key(), it.Value() )

	}


	return batch
}


func (cache *aCache) Update( peerId string, node *Node ) error {

	exist, err := AvdbComm.CacheHas( cache.source, cache.cdb, []byte(peerId) )

	if err != nil {
		return err
	}

	if !exist {
		return leveldb.ErrNotFound
	}

	return cache.cdb.Put([]byte(peerId), node.Encode(), nil)

}


func (cache *aCache) Insert( peerId string, node *Node ) error {

	return cache.cdb.Put([]byte(peerId), node.Encode(), nil)

}


func (cache *aCache) Del( peerId string ) {

	_ = cache.cdb.Delete([]byte(peerId), nil)

	cache.delKeys = append(cache.delKeys, []byte(peerId))

}

func (cache *aCache) GetFirst() *Node {

	it := cache.source.NewIterator( &util.Range{nil,nil}, nil )

	defer it.Release()

	var maxnd *Node

	for it.Next() {

		nd := &Node{}

		if err := nd.Decode(it.Value()); err != nil {
			continue
		}

		if maxnd == nil {

			maxnd = nd
			continue

		} else {

			if nd.Votes > maxnd.Votes {
				maxnd = nd
			}

		}

	}

	return maxnd
}

func (cache *aCache) GetLatest() *Node {

	it := cache.source.NewIterator( &util.Range{nil,nil}, nil )

	defer it.Release()

	var minnd *Node

	for it.Next() {

		nd := &Node{}

		if err := nd.Decode(it.Value()); err != nil {
			continue
		}

		if minnd == nil {

			minnd = nd
			continue

		} else {

			if nd.Votes < minnd.Votes {
				minnd = nd
			}

		}

	}

	return minnd

}

func (cache *aCache) TotalSum() uint64 {

	size, err := cache.source.SizeOf([]util.Range{{nil,nil}})
	if err != nil {
		return 0
	}

	return uint64(size.Sum())
}