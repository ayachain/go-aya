package transaction

import (
	"fmt"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"sync"
)

type aTransactions struct {
	TransactionAPI
	*mfs.Directory

	mfsstorage storage.Storage
	rawdb *leveldb.DB
	RWLocker sync.RWMutex
}

func CreateServices( mdir *mfs.Directory, rdOnly bool ) TransactionAPI {

	api := &aTransactions{
		Directory:mdir,
	}

	api.rawdb, api.mfsstorage = AVdbComm.OpenExistedDB(mdir, DBPath, rdOnly)

	return api
}

func (txs *aTransactions) DBKey()	string {
	return DBPath
}

func (txs *aTransactions) GetTxByHash( hash EComm.Hash ) (*Transaction, error) {

	tx := &Transaction{}

	val, err := txs.rawdb.Get(hash.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	if err := tx.Decode(val); err != nil {
		return nil, fmt.Errorf("%v can't found transaction", hash.String())
	}

	return tx, nil
}


func (txs *aTransactions) GetTxByHashBs( hsbs []byte ) (*Transaction, error) {

	hash := EComm.BytesToHash(hsbs)

	return txs.GetTxByHash(hash)
}


func (api *aTransactions) OpenVDBTransaction() (*leveldb.Transaction, *sync.RWMutex, error) {

	tx, err := api.rawdb.OpenTransaction()
	if err != nil {
		return nil, nil, err
	}

	return tx, &api.RWLocker, nil
}

func (api *aTransactions) Close() {

	api.RWLocker.Lock()
	defer api.RWLocker.Unlock()

	_ = api.rawdb.Close()
	_ = api.mfsstorage.Close()
	_ = api.Flush()
}