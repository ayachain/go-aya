package txutils

import (
	"github.com/ayachain/go-aya/vdb/im"
	EComm "github.com/ethereum/go-ethereum/common"
)

func MakeTransferAvail( from EComm.Address, to EComm.Address, value uint64, tid uint64 ) *im.Transaction {

	tx := &im.Transaction{}
	tx.From = from.Bytes()
	tx.To = to.Bytes()
	tx.Value = value
	tx.Data = ""
	tx.Tid = tid
	tx.Type = im.TransactionType_Normal

	return tx

}