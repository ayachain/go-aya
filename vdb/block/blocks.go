package block

import (
	"encoding/binary"
	"errors"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"strings"
	"sync"
)

var(
	LatestPosBlockIdxKey 	= []byte("LatestPosBlockIndex")
)

type aBlocks struct {

	reader
	*mfs.Directory

	headAPI AIndexes.IndexesServices

	ldb *leveldb.DB
	mfsstorage *ADB.MFSStorage
	dbSnapshot *leveldb.Snapshot
	snLock sync.RWMutex

}

func CreateServices( mdir *mfs.Directory, hapi AIndexes.IndexesServices ) Services {

	var err error

	api := &aBlocks{
		Directory:mdir,
		headAPI:hapi,
	}

	api.ldb, api.mfsstorage, err = AVdbComm.OpenExistedDB( mdir, DBPath )
	if err != nil {
		panic(err)
	}

	api.dbSnapshot, err = api.ldb.GetSnapshot()
	if err != nil {
		_ = api.ldb.Close()
		log.Error(err)
		return nil
	}

	return api
}

func (blks *aBlocks) GetLatestPosBlockIndex() uint64 {

	blks.snLock.RLock()
	defer blks.snLock.RUnlock()

	if exist, err := blks.dbSnapshot.Has(LatestPosBlockIdxKey, nil); err != nil {

		panic(err)

	} else if !exist {

		return 0

	} else {

		if bs, err := blks.dbSnapshot.Get(LatestPosBlockIdxKey, nil); err != nil {

			return 0

		} else {

			return binary.BigEndian.Uint64(bs)

		}

	}

}

func (blks *aBlocks) GetBlocks( hashOrIndex...interface{} ) ([]*Block, error) {

	blks.snLock.RLock()
	defer blks.snLock.RUnlock()

	var blist []*Block

	for _, v := range hashOrIndex {

		var bhash EComm.Hash

		switch v.(type) {

		case string:

			switch strings.ToLower(v.(string)) {
			case BlockNameLatest:

				hd, err := blks.headAPI.GetLatest()
				if err != nil {
					return nil, err
				}
				bhash = hd.Hash


			case BlockNameGen:

				hd, err := blks.headAPI.GetIndex(0)
				if err != nil {
					return nil, err
				}
				bhash = hd.Hash
			}

		case uint64:

			hd, err := blks.headAPI.GetIndex(v.(uint64))
			if err != nil {
				return nil, err
			}
			bhash = hd.Hash

		case EComm.Hash:
			bhash = v.(EComm.Hash)

		default:
			return nil, errors.New("input params must be a index(uint64) or cid object")
		}

		dbval, err := blks.dbSnapshot.Get( bhash.Bytes(), nil )
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

func (blks *aBlocks) NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( blks.dbSnapshot, blks.headAPI )
}

func (blks *aBlocks) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := blks.ldb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (blks *aBlocks) Shutdown() error {

	blks.snLock.Lock()
	defer blks.snLock.Unlock()

	if blks.dbSnapshot != nil {
		blks.dbSnapshot.Release()
	}

	//if err := blks.mfsstorage.Close(); err != nil {
	//	return err
	//}

	if err := blks.ldb.Close(); err != nil {
		return err
	}

	return nil
}

func (blks *aBlocks) Close() {
	_ = blks.Shutdown()
}

func (api *aBlocks) UpdateSnapshot() error {

	api.snLock.Lock()
	defer api.snLock.Unlock()

	if api.dbSnapshot != nil {
		api.dbSnapshot.Release()
	}

	var err error
	api.dbSnapshot, err = api.ldb.GetSnapshot()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (api *aBlocks) SyncCache() error {

	if err := api.ldb.CompactRange(util.Range{nil,nil}); err != nil {
		log.Error(err)
	}

	return api.mfsstorage.Flush()
}