package receipt

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
)

const DBPath = "/db/receipts"

type ReceiptsAPI interface {
	AVdbComm.VDBSerices

	GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error)
	HasTransactionReceipt( txhs EComm.Hash ) bool
}