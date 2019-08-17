package transaction

import (
	"context"
	"encoding/binary"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aCache struct {

	MergeWriter

	sourceReader *aTransactions

	cdb *leveldb.DB
}

func newWriter( sreader *aTransactions ) (MergeWriter, error) {

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

func (cache *aCache) Put(tx *Transaction, bidx uint64) {

	lidx, err := cache.sourceReader.idxs.GetLatest()
	if err != nil {
		panic(err)
	}

	dbroot, err, cls := AvdbComm.GetDBRoot(context.TODO(), cache.sourceReader.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		countKey := append(TxOutCountPrefix, tx.From.Bytes()... )

		exist, err := AvdbComm.CacheHas(db, cache.cdb, countKey)
		if err != nil {
			panic(err)
		}

		var txcount uint64 = 0
		if exist {

			cbs, err := AvdbComm.CacheGet(db, cache.cdb, countKey)
			if err != nil {
				panic(err)
			}

			txcount = binary.BigEndian.Uint64(cbs)
		}

		key := append(tx.GetHash256().Bytes(), AvdbComm.BigEndianBytes(bidx)...)

		tx.BlockIndex = bidx

		if err := cache.cdb.Put(key, tx.Encode(), AvdbComm.WriteOpt); err != nil {
			panic(err)
		}

		// Block index resolve
		tx.BlockIndex = 0

		txcount ++

		if err := cache.cdb.Put( countKey, AvdbComm.BigEndianBytes(txcount), AvdbComm.WriteOpt ); err != nil {
			panic(err)
		}

		// Write from tx history
		fromTxTotalKey := append(TxTotalCountPrefix, tx.From.Bytes()...)
		toTxTotalKey := append(TxTotalCountPrefix, tx.To.Bytes()...)

		var (
			fromTxTotal uint64 = 0
			toTxTotal uint64 = 0
		)

		// from tx total key exist
		fttke, err := AvdbComm.CacheHas(db, cache.cdb, fromTxTotalKey)
		if err != nil {

			panic(err)

		} else {

			if fttke {

				cbs, err := AvdbComm.CacheGet(db, cache.cdb, fromTxTotalKey)
				if err != nil {
					panic(err)
				}

				fromTxTotal = binary.BigEndian.Uint64(cbs)

			}

			hkey := append(TxHistoryPrefix, tx.From.Bytes()... )
			hkey = append(hkey, AvdbComm.BigEndianBytes(fromTxTotal)...)

			fromTxTotal ++

			// Update total count
			if err := cache.cdb.Put( fromTxTotalKey, AvdbComm.BigEndianBytes(fromTxTotal), AvdbComm.WriteOpt ); err != nil {
				panic(err)
			}

			// insert history indexes
			if err := cache.cdb.Put( hkey, key, AvdbComm.WriteOpt ); err != nil {
				panic(err)
			}

		}

		// to tx total key exist
		tttke, err :=  AvdbComm.CacheHas(db, cache.cdb, toTxTotalKey)
		if err != nil {

			panic(err)

		} else {

			if tttke {

				cbs, err := AvdbComm.CacheGet(db, cache.cdb, toTxTotalKey)
				if err != nil {
					panic(err)
				}

				toTxTotal = binary.BigEndian.Uint64(cbs)
			}

			hkey := append(TxHistoryPrefix, tx.To.Bytes()... )
			hkey = append(hkey, AvdbComm.BigEndianBytes(toTxTotal)...)

			toTxTotal ++

			// Update total count
			if err := cache.cdb.Put( toTxTotalKey, AvdbComm.BigEndianBytes(toTxTotal), AvdbComm.WriteOpt ); err != nil {
				panic(err)
			}

			// insert history indexes
			if err := cache.cdb.Put( hkey, key, AvdbComm.WriteOpt ); err != nil {
				panic(err)
			}

		}

		return nil

	}, DBPath ); err != nil {
		log.Error(err)
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
func (cache *aCache) GetTxByHash( hash EComm.Hash, idx ... *indexes.Index ) (*Transaction, error) {
	return cache.sourceReader.GetTxByHash(hash)
}

func (cache *aCache) GetTxByHashBs( hsbs []byte, idx ... *indexes.Index ) (*Transaction, error) {
	return cache.sourceReader.GetTxByHashBs(hsbs)
}

func (cache *aCache) GetTxCount( address EComm.Address, idx ... *indexes.Index ) (uint64, error) {
	return cache.sourceReader.GetTxCount(address)
}

func (cache *aCache) GetHistoryHash( address EComm.Address, offset uint64, size uint64, idx ... *indexes.Index ) []EComm.Hash {
	return cache.sourceReader.GetHistoryHash(address, offset, size)
}

func (cache *aCache) GetHistoryContent( address EComm.Address, offset uint64, size uint64, idx ... *indexes.Index ) ([]*Transaction, error) {
	return cache.sourceReader.GetHistoryContent(address, offset, size)
}