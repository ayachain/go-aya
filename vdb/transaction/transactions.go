package transaction

import (
	"fmt"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
)

type aTransactions struct {
	TransactionAPI
	*mfs.Directory
	rawdb *leveldb.DB
}

func CreateServices( mdir *mfs.Directory ) TransactionAPI {

	api := &aTransactions{
		Directory:mdir,
	}

	api.rawdb = AVdbComm.OpenExistedDB(mdir, DBPath)

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