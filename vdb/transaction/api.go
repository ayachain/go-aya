package transaction

import (
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

const DBPath = "/db/header"

type TransactionAPI interface {

	DBKey()	string

	GetTxByHash( hash EComm.Hash ) (*Transaction, error)

	GetTxByHashBs( hsbs []byte ) (*Transaction, error)

	NewBlockTxIterator( bindex uint64 ) iterator.Iterator
}
