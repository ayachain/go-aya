package transaction

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
)

const DBPath = "/transactions"

type reader interface {

	GetTxByHash( hash EComm.Hash ) (*Transaction, error)

	GetTxByHashBs( hsbs []byte ) (*Transaction, error)

	GetTxCount( address EComm.Address ) (uint64, error)

	GetHistoryHash( address EComm.Address, offset uint64, size uint64) []EComm.Hash

	GetHistoryContent( address EComm.Address, offset uint64, size uint64) ([]*Transaction, error)
}


type writer interface {
	Put(tx *Transaction, bidx uint64)
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