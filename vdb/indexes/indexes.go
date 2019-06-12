package indexes

import (
	"github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

type aIndexes struct {

	IndexesAPI

	RWLocker sync.RWMutex

	rawdb *leveldb.DB
}

func CreateServices( chanid string ) IndexesAPI {

	path := "~/.aya/chain/indexes/" + chanid

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		panic(err)
	}

	api := &aIndexes{
		rawdb:db,
	}

	return api
}

func ( i *aIndexes ) PutIndex( index *Index ) error {

	key := common.BigEndianBytes(index.BlockIndex)
	value := index.Encode()

	if err := i.rawdb.Put(key, value, nil); err != nil {
		return err
	}

	return nil

}

func ( i *aIndexes ) PutIndexBy( num uint64, bhash EComm.Hash, ci cid.Cid ) error {

	key := common.BigEndianBytes(num)
	value := (&Index{
		BlockIndex:num,
		Hash:bhash,
		FullCID:ci,
	}).Encode()

	if err := i.rawdb.Put(key, value, nil); err != nil {
		return err
	}

	return nil

}

func ( i *aIndexes ) Close() error {

	return i.rawdb.Close()

}

func ( i *aIndexes ) GetIndex( blockNumber uint64 ) (*Index, error) {

	key := common.BigEndianBytes(blockNumber)

	bs, err := i.rawdb.Get(key, nil)
	if err != nil {
		return nil, err
	}

	idx := &Index{}
	if err := idx.Decode(bs); err != nil {
		return nil, err
	}

	return idx,nil

}

