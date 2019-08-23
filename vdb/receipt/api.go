package receipt

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
)

const DBPath = "/receipts"

type reader interface {

	GetTransactionReceipt( txhs EComm.Hash, idx... *indexes.Index ) (*im.Receipt, error)

	HasTransactionReceipt( txhs EComm.Hash, idx... *indexes.Index ) bool
}


type writer interface {
	Put( txhs EComm.Hash, bindex uint64, receipt *im.Receipt )
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