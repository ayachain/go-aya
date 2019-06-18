package transaction

import (
	"encoding/binary"
	"fmt"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)


var TxCountPrefix = []byte("TxCount_")

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

func (txs *aTransactions) NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( txs.rawdb )
}

func (txs *aTransactions) Close() {

	_ = txs.rawdb.Close()
	_ = txs.mfsstorage.Close()
	_ = txs.Flush()

}

func (txs *aTransactions) Shutdown() error {

	_ = txs.rawdb.Close()
	_ = txs.mfsstorage.Close()

	return txs.Flush()
}

func (txs *aTransactions) GetTxCount( address EComm.Address ) (uint64, error) {

	key := append(TxCountPrefix, address.Bytes()... )

	v, err := txs.rawdb.Get(key, nil)

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

	tx := &Transaction{}

	st := append(hash.Bytes(), AVdbComm.BigEndianBytes(0)... )

	it := txs.rawdb.NewIterator(&util.Range{Start:st, Limit:nil}, nil)
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

	tx, err := txs.rawdb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}
