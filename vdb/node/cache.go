package node

import (
	"fmt"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/indexes"
	"github.com/golang/protobuf/proto"
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

func (cache *aCache) InsertBootstrapNodes( nds []*im.Node ) {

	for i, v := range nds {

		key := []byte( fmt.Sprintf("%v%d", im.NodeType_Super, i) )

		value, err := proto.Marshal(v)
		if err != nil {
			panic(err)
		}

		if err := cache.cdb.Put( []byte(v.PeerID), value, AvdbComm.WriteOpt ); err != nil {
			panic(err)
		}

		if err := cache.cdb.Put( key, []byte(v.PeerID), AvdbComm.WriteOpt ); err != nil {
			panic(err)
		}

	}
}

func (cache *aCache) Insert( peerId string, node *im.Node ) error {

	if exist, err := cache.cdb.Has([]byte(peerId), nil); err == nil && exist {
		return ErrNodeAlreadyExist
	}

	if node.Type == im.NodeType_Super {

		superNodes := cache.sourceReader.GetSuperNodeList()

		superNodes = append(superNodes, node)

		sort.SliceSorter(superNodes, func(i, j int) bool {
			return superNodes[i].Votes < superNodes[j].Votes
		})

		if len(superNodes) > 21 {
			superNodes = superNodes[:21]
		}

		for i, v := range superNodes {

			key := []byte( fmt.Sprintf("%v%d", im.NodeType_Super, i) )

			bs, err := proto.Marshal(v)
			if err != nil {
				panic(err)
			}

			if err := cache.cdb.Put( key, bs, AvdbComm.WriteOpt ); err != nil {
				return err
			}

		}

		return nil

	} else {

		bs, err := proto.Marshal(node)
		if err != nil {
			panic(err)
		}

		return cache.cdb.Put([]byte(peerId), bs, AvdbComm.WriteOpt)
	}

}

func (cache *aCache) Del( peerId string ) {

	nd, err := cache.sourceReader.GetNodeByPeerId(peerId)
	if err != nil {
		return
	}

	if nd.Type == im.NodeType_Super {

		n := 0

		superNodes := cache.sourceReader.GetSuperNodeList()

		for _, v := range superNodes {

			if v.PeerID == peerId {

				bs, err := proto.Marshal(v)
				if err != nil {
					panic(err)
				}

				_ = cache.cdb.Put( []byte(fmt.Sprintf("%v%d", im.NodeType_Super, n)), bs, AvdbComm.WriteOpt )
			}

			n ++
		}

		delkey := []byte(fmt.Sprintf("%v%d", im.NodeType_Super, n - 1))

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

func (cache *aCache) GetNodeByPeerId( peerId string, idx ... *indexes.Index ) (*im.Node, error) {
	return cache.sourceReader.GetNodeByPeerId(peerId)
}

func (cache *aCache) GetSuperMaterTotalVotes( idx ... *indexes.Index ) uint64 {
	return cache.sourceReader.GetSuperMaterTotalVotes()
}

func (cache *aCache) GetSuperNodeList( idx ... *indexes.Index ) []*im.Node {
	return cache.sourceReader.GetSuperNodeList()
}

func (cache *aCache) GetSuperNodeCount( idx ... *indexes.Index ) int64 {
	return cache.sourceReader.GetSuperNodeCount()
}