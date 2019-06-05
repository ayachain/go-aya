package receipt

import EComm "github.com/ethereum/go-ethereum/common"

type ReceiptsAPI interface {

	DBKey()	string

	GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error)

}