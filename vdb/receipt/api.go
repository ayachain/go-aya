package receipt

import EComm "github.com/ethereum/go-ethereum/common"

type ReceiptsAPI interface {
	GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error)
}