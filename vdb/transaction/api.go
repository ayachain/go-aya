package transaction

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

const DBPath = "/db/transactions"

type reader interface {

	GetTxByHash( hash EComm.Hash ) (*Transaction, error)

	GetTxByHashBs( hsbs []byte ) (*Transaction, error)

	NewBlockTxIterator( bindex uint64 ) iterator.Iterator
}


type writer interface {

}


type Services interface {

	AVdbComm.VDBSerices
	reader
}


type Caches interface {

	AVdbComm.VDBCacheServices

	reader
	writer
}