package txutils

import (
	EComm "github.com/ethereum/go-ethereum/common"
	ATransaction "github.com/ayachain/go-aya/vdb/transaction"
)

func MakeTransferAvail( from EComm.Address, to EComm.Address, value uint64, tid uint64 ) *ATransaction.Transaction {

	tx := &ATransaction.Transaction{}
	tx.From = from
	tx.To = to
	tx.Value = value
	tx.Data = nil
	tx.Tid = tid

	return tx

}