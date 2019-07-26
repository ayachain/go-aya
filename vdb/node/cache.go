package node

import (
	"fmt"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go4.org/sort"
)


var (
	ErrNodeAlreadyExist = errors.New("insert target node already exist.")
)

type aCache struct {

	writer

	headAPI AIndexes.IndexesServices
	source *leveldb.Snapshot
	cdb *leveldb.DB

	delKeys [][]byte
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

func (cache *aCache) GetSnapshot() *leveldb.Snapshot {
	return cache.source
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


//func (cache *aCache) Update( peerId string, node *Node ) error {
//
//	exist, err := AvdbComm.CacheHas( cache.source, cache.cdb, []byte(peerId) )
//
//	if err != nil {
//		return err
//	}
//
//	if !exist {
//		return leveldb.ErrNotFound
//	}
//
//	return cache.cdb.Put([]byte(peerId), node.Encode(), AvdbComm.WriteOpt)
//
//}

func (cache *aCache) GetSuperMaterTotalVotes() uint64 {

	var total uint64

	it := cache.source.NewIterator( util.BytesPrefix( []byte(NodeTypeSuper) ), nil )

	defer it.Release()

	for it.Next() {

		perrId := it.Value()

		bs, err := cache.source.Get(perrId, nil)

		if err != nil {
			panic(err)
		}

		nd := &Node{}

		if err := nd.Decode(bs); err != nil {
			total += nd.Votes
		}
	}

	return total
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

func (cache *aCache) GetSuperNodeList() []*Node {

	var rets[] *Node

	it := cache.source.NewIterator( util.BytesPrefix( []byte(NodeTypeSuper) ), nil )

	defer it.Release()

	for it.Next() {

		perrId := it.Value()

		bs, err := cache.source.Get(perrId, nil)

		if err != nil {
			panic(err)
		}

		nd := &Node{}

		if err := nd.Decode(bs); err != nil {
			rets = append(rets, nd)
		}
	}

	sort.SliceSorter(rets, func(i, j int) bool {
		return rets[i].Votes < rets[j].Votes
	})

	return rets
}

func (cache *aCache) Insert( peerId string, node *Node ) error {

	if exist, err := AvdbComm.CacheHas( cache.source, cache.cdb, []byte(peerId) ); err == nil || exist {

		return ErrNodeAlreadyExist

	}

	if node.Type == NodeTypeSuper {

		superNodes := cache.GetSuperNodeList()

		superNodes = append(superNodes, node)

		sort.SliceSorter(superNodes, func(i, j int) bool {
			return superNodes[i].Votes < superNodes[j].Votes
		})

		superNodes = superNodes[:21]

		for i, v := range superNodes {

			key := []byte( fmt.Sprintf("%v%d", NodeTypeSuper, i) )

			if err := cache.cdb.Put( key, v.Encode(), AvdbComm.WriteOpt ); err != nil {
				return err
			}

		}

		return nil

	} else {

		return cache.cdb.Put([]byte(peerId), node.Encode(), AvdbComm.WriteOpt)

	}



}


func (cache *aCache) Del( peerId string ) {

	nd, err := cache.GetNodeByPeerId(peerId)
	if err != nil {
		return
	}

	if nd.Type == NodeTypeSuper {

		n := 0

		superNodes := cache.GetSuperNodeList()

		for _, v := range superNodes {

			if v.PeerID == peerId {
				_ = cache.cdb.Put( []byte(fmt.Sprintf("%v%d", NodeTypeSuper, n)), v.Encode(), AvdbComm.WriteOpt )
			}

			n ++
		}

		delkey := []byte(fmt.Sprintf("%v%d", NodeTypeSuper, n - 1))

		_ = cache.cdb.Delete(delkey, nil)

		cache.delKeys = append(cache.delKeys, delkey)
	}

	_ = cache.cdb.Delete([]byte(peerId), nil)

	cache.delKeys = append(cache.delKeys, []byte(peerId))
}