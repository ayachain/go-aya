package transaction

import (
	"bytes"
	"fmt"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type aTransactions struct {
	TransactionAPI
	rawdb *leveldb.DB
}


func CreateTransactionsAPI( rootref *mfs.Root ) TransactionAPI {

	api := &aTransactions{}

	api.rawdb = AvdbComm.OpenStandardDB(rootref, DBPath)

	return api
}


func (txs *aTransactions) GetTxByHash( hash EComm.Hash ) (*Transaction, error) {

	stbuf := bytes.NewBuffer( hash.Bytes() )
	stbuf.Write( AvdbComm.BigEndianBytes(0) )

	edbuf := bytes.NewBuffer( hash.Bytes() )
	edbuf.Write( AvdbComm.BigEndianBytes(-1) )

	it := txs.rawdb.NewIterator( &util.Range{
		Start : stbuf.Bytes(),
		Limit : edbuf.Bytes(),
	}, nil)

	defer it.Release()

	if it.Next() {

		tx := &Transaction{}

		if err := tx.Decode(it.Value()); err != nil {
			return nil, err
		}

		return tx, nil
	}

	return nil, fmt.Errorf("%v can't found transaction", hash.String())
}


func (txs *aTransactions) GetTxByHashBs( hsbs []byte ) (*Transaction, error) {

	hash := EComm.BytesToHash(hsbs)

	return txs.GetTxByHash(hash)
}