package common

import (
	"encoding/binary"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
)

const(

	TxInsufficientFunds = "Insufficient funds"
	TxConfirm = "Confirm"
	TxOverrided = "Transaction Override"
	TxUnsupportTransactionType = "Unsupported transaction type"

)


type Record struct {

	AVdbComm.RawDBCoder

	LastUpdateTime 	uint16
	VoteSum			uint64
}

func (r *Record) Encode() []byte {

	bs := make([]byte, 2 + 8)

	copy(bs[0:2], AVdbComm.BigEndianBytesUint16(r.LastUpdateTime))
	copy(bs[3:7], AVdbComm.BigEndianBytesUint16(r.LastUpdateTime))

	//bs[3:7] = AVdbComm.BigEndianBytes(r.VoteSum)

	return bs
}

func (r *Record) Decode(bs []byte) error {

	r.LastUpdateTime = binary.BigEndian.Uint16(bs[0:2])
	r.VoteSum = binary.BigEndian.Uint64(bs[3:])

	return nil
}