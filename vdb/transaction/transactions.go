package transaction

import (
	"encoding/binary"
	"fmt"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

var TxCountPrefix = []byte("TxCount_")

type aTransactions struct {

	Services
	*mfs.Directory

	mfsstorage *ADB.MFSStorage
	ldb *leveldb.DB
	dbSnapshot *leveldb.Snapshot
	snLock sync.RWMutex
}

func CreateServices( mdir *mfs.Directory ) Services {

	var err error

	api := &aTransactions{
		Directory:mdir,
	}

	api.ldb, api.mfsstorage = AVdbComm.OpenExistedDB(mdir, DBPath)

	api.dbSnapshot, err = api.ldb.GetSnapshot()
	if err != nil {
		_ = api.ldb.Close()
		log.Error(err)
		return nil
	}

	return api
}


func (txs *aTransactions) NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( txs.dbSnapshot )
}


func (txs *aTransactions) Close() {
	_ = txs.Shutdown()
}


func (txs *aTransactions) Shutdown() error {

	txs.snLock.Lock()
	defer txs.snLock.Unlock()

	if txs.dbSnapshot != nil {
		txs.dbSnapshot.Release()
	}

	//if err := txs.mfsstorage.Close(); err != nil {
	//	return err
	//}

	if err := txs.ldb.Close(); err != nil {
		log.Error(err)
		return err
	}

	return nil
}


func (txs *aTransactions) GetTxCount( address EComm.Address ) (uint64, error) {

	txs.snLock.RLock()
	defer txs.snLock.RUnlock()

	key := append(TxCountPrefix, address.Bytes()... )

	v, err := txs.dbSnapshot.Get(key, nil)

	if err != nil {

		if err == leveldb.ErrNotFound {
			return 0, nil
		} else {
			return 0, err
		}

	}

	return binary.BigEndian.Uint64(v), nil
}


func (txs *aTransactions) GetTxByHash( hash EComm.Hash ) (*Transaction, error) {

	txs.snLock.RLock()
	defer txs.snLock.RUnlock()

	tx := &Transaction{}

	it := txs.dbSnapshot.NewIterator( util.BytesPrefix(hash.Bytes()) , nil)
	defer it.Release()

	if it.Next() {

		if err := tx.Decode(it.Value()); err != nil {
			return nil, fmt.Errorf("%v can't found transaction", hash.String())
		}

	} else {
		return nil, fmt.Errorf("%v can't found transaction", hash.String())
	}

	return tx, nil
}


func (txs *aTransactions) GetTxByHashBs( hsbs []byte ) (*Transaction, error) {

	hash := EComm.BytesToHash(hsbs)

	return txs.GetTxByHash(hash)
}


func (txs *aTransactions) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := txs.ldb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}


func (api *aTransactions) UpdateSnapshot() error {

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

func (api *aTransactions) SyncCache() error {

	if err := api.ldb.CompactRange(util.Range{nil,nil}); err != nil {
		log.Error(err)
	}

	return api.mfsstorage.Flush()
}