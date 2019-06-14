package transaction

import (
	"fmt"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aTransactions struct {

	Services
	*mfs.Directory

	mfsstorage storage.Storage
	rawdb *leveldb.DB
}

func CreateServices( mdir *mfs.Directory, rdOnly bool ) Services {

	api := &aTransactions{
		Directory:mdir,
	}

	api.rawdb, api.mfsstorage = AVdbComm.OpenExistedDB(mdir, DBPath, rdOnly)

	return api
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

func (txs *aTransactions) Close() {

	_ = txs.rawdb.Close()
	_ = txs.mfsstorage.Close()
	_ = txs.Flush()
}

func (txs *aTransactions) NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( txs.rawdb )
}

func (txs *aTransactions) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := txs.rawdb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (txs *aTransactions) Shutdown() error {

	_ = txs.rawdb.Close()
	_ = txs.mfsstorage.Close()

	return txs.Flush()
}