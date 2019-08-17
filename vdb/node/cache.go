package node

import (
	"fmt"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"go4.org/sort"
)

var (
	ErrNodeAlreadyExist = errors.New("insert target node already exist.")
)

type aCache struct {

	writer

	sourceReader reader

	cdb *leveldb.DB

	delKeys [][]byte
}

func newWriter( sread reader ) (MergeWriter, error) {

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

func (cache *aCache) MergerBatch() *leveldb.Batch {

	batch := &leveldb.Batch{}

	it := cache.cdb.NewIterator(nil, nil)
	defer it.Release()

	for _, delk := range cache.delKeys {
		batch.Delete(delk)
	}

	for it.Next() {
		batch.Put( it.Key(), it.Value() )
	}

	return batch
}

func (cache *aCache) InsertBootstrapNodes( nds []Node ) {

	for i, v := range nds {

		key := []byte( fmt.Sprintf("%v%d", NodeTypeSuper, i) )

		value := v.Encode()

		if err := cache.cdb.Put( []byte(v.PeerID), value, AvdbComm.WriteOpt ); err != nil {
			panic(err)
		}

		if err := cache.cdb.Put( key, []byte(v.PeerID), AvdbComm.WriteOpt ); err != nil {
			panic(err)
		}

	}
}

func (cache *aCache) Insert( peerId string, node *Node ) error {

	if exist, err := cache.cdb.Has([]byte(peerId), nil); err == nil && exist {
		return ErrNodeAlreadyExist
	}

	if node.Type == NodeTypeSuper {

		superNodes := cache.sourceReader.GetSuperNodeList()

		superNodes = append(superNodes, node)

		sort.SliceSorter(superNodes, func(i, j int) bool {
			return superNodes[i].Votes < superNodes[j].Votes
		})

		if len(superNodes) > 21 {
			superNodes = superNodes[:21]
		}

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

	nd, err := cache.sourceReader.GetNodeByPeerId(peerId)
	if err != nil {
		return
	}

	if nd.Type == NodeTypeSuper {

		n := 0

		superNodes := cache.sourceReader.GetSuperNodeList()

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

func (cache *aCache) Close() {
	_ = cache.cdb.Close()
}

/// Reader mock impl
func (cache *aCache) DoRead( readingFunc ADB.ReadingFunc, idx ... *indexes.Index ) error {
	return cache.sourceReader.DoRead( readingFunc )
}

func (cache *aCache) GetNodeByPeerId( peerId string, idx ... *indexes.Index ) (*Node, error) {
	return cache.sourceReader.GetNodeByPeerId(peerId)
}

func (cache *aCache) GetSuperMaterTotalVotes( idx ... *indexes.Index ) uint64 {
	return cache.sourceReader.GetSuperMaterTotalVotes()
}

func (cache *aCache) GetSuperNodeList( idx ... *indexes.Index ) []*Node {
	return cache.sourceReader.GetSuperNodeList()
}

func (cache *aCache) GetSuperNodeCount( idx ... *indexes.Index ) int64 {
	return cache.sourceReader.GetSuperNodeCount()
}