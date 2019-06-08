package transaction

import (
	"bytes"
	"errors"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	AvdbTx "github.com/ayachain/go-aya/vdb/transaction"
)

const MessagePrefix = byte('t')

var (
	ErrMsgPrefix = errors.New("not a chain info message")
)


type MsgRawTx struct {
	AvdbComm.RawDBCoder
	AvdbTx.Transaction
}


func (msg *MsgRawTx) Prefix() byte{
	return MessagePrefix
}


func (msg *MsgRawTx) Encode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( msg.Transaction.Encode() )

	return buff.Bytes()
}


func (msg *MsgRawTx) Decode(bs []byte) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return msg.Transaction.Decode(bs[1:])
}

