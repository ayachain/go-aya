package block

import (
	"bytes"
	"errors"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AvdbComm"github.com/ayachain/go-aya/vdb/common"
)

// 1 byte prefix + 4 * Hash(32Byte) = 129 byte
const MessagePrefix = byte('b')

var (
	ErrMsgPrefix = errors.New("not a chain info message")
)

type MsgRawBlock struct {
	AvdbComm.RawDBCoder
	ABlock.Block
}

func (msg *MsgRawBlock) Encode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})
	buff.Write( msg.Block.Encode() )

	return buff.Bytes()
}

func (msg *MsgRawBlock) Decode(bs []byte) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return msg.Block.Decode(bs[1:])
}

