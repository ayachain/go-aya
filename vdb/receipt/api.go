package receipt

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
)

const DBPath = "/db/receipts"


type reader interface {
	GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error)
	HasTransactionReceipt( txhs EComm.Hash ) bool
}


type writer interface {

	Put( txhs EComm.Hash, bindex uint64, receipt []byte )
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