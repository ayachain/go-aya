package transaction

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
)

const DBPath = "/transactions"

type reader interface {

	GetTxByHash( hash EComm.Hash, idx... *indexes.Index ) (*im.Transaction, error)

	GetTxByHashBs( hsbs []byte, idx... *indexes.Index ) (*im.Transaction, error)

	GetTxCount( address EComm.Address, idx... *indexes.Index ) (uint64, error)

	GetHistoryHash( address EComm.Address, offset uint64, size uint64, idx... *indexes.Index) []EComm.Hash

	GetHistoryContent( address EComm.Address, offset uint64, size uint64, idx... *indexes.Index) ([]*im.Transaction, error)
}


type writer interface {

	Put(tx *im.Transaction, bidx uint64)
}

type Services interface {
	AVdbComm.VDBSerices
	reader
}

type MergeWriter interface {
	AVdbComm.VDBCacheServices
	reader
	writer
}