package transaction

import (
	"github.com/ayachain/go-aya/vdb/im"
)

const MessagePrefix = byte('t')

var StaticCostMapping = map[im.TransactionType]uint64{
	im.TransactionType_Normal 		: 150000,
	im.TransactionType_Unrecorded 	: 100000,
}

type ConfirmTx struct {

	im.Transaction

	Time 			uint32			`json:"Time"`
}