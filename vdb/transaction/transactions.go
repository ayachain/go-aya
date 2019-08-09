package transaction

import (
	"encoding/binary"
	"fmt"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

var (

	TxOutCountPrefix = []byte("TOC_")

	TxTotalCountPrefix = []byte("TTC_")

	TxHistoryPrefix = []byte("TH_")

)

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

	api.ldb, api.mfsstorage, err = AVdbComm.OpenExistedDB(mdir, DBPath)
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

	key := append(TxOutCountPrefix, address.Bytes()... )

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


func (api *aTransactions) GetHistoryHash( address EComm.Address, offset uint64, size uint64) []EComm.Hash {

	itkey := append(TxHistoryPrefix, address.Bytes()...)

	it := api.ldb.NewIterator( util.BytesPrefix(itkey), nil )

	var hashs []EComm.Hash

	for it.Next() {
		hashs = append(hashs, EComm.BytesToHash(it.Value()))
	}

	return hashs

}

func (api *aTransactions) GetHistoryContent( address EComm.Address, offset uint64, size uint64) ([]*Transaction, error) {

	itkey := append(TxHistoryPrefix, address.Bytes()...)

	it := api.ldb.NewIterator( util.BytesPrefix(itkey), nil )
	defer it.Release()

	var txs []*Transaction

	if offset > 0 {

		seekKey := append(itkey, AVdbComm.BigEndianBytes(offset)...)

		if !it.Seek(seekKey) {
			return nil, errors.New("seek history key expected ")
		}
	}

	s := uint64(0)

	for it.Next() {

		if s >= size - 1 {
			break
		}

		bs, err := api.ldb.Get(it.Value(), nil)
		if err != nil {
			return nil, err
		}

		tx := &Transaction{}

		if err := tx.Decode(bs); err != nil {
			return nil, err
		}

		txs = append(txs, tx)

		s ++
	}

	return txs, nil
}