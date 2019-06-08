package chaininfo

import (
	"bytes"
	"errors"
	AvdbComm"github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
)

// 1 byte prefix + 4 * Hash(32Byte) = 129 byte
const MessagePrefix = byte('i')
const MessageSize = 129

var (
	ErrMsgPrefix    = errors.New("not a chain info message")
	ErrLenLess 		= errors.New("message content len is expected")
)

type MsgRawChainInfo struct {
	AvdbComm.RawDBCoder
	GenBlock 	EComm.Hash	 	`json:"G"`
	VDBRoot		EComm.Hash		`json:"V"`
	Header		EComm.Hash 		`json:"H"`
	Block		EComm.Hash		`json:"B"`
}

func (msg *MsgRawChainInfo) Prefix() byte{
	return MessagePrefix
}

func (msg *MsgRawChainInfo) Encode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( msg.GenBlock.Bytes() )
	buff.Write( msg.VDBRoot.Bytes() )
	buff.Write( msg.Header.Bytes() )
	buff.Write( msg.Block.Bytes() )

	return buff.Bytes()
}

func (msg *MsgRawChainInfo) Decode(bs []byte) error {

	if len(bs) != MessageSize {
		return ErrMsgPrefix
	}

	if bs[0] != MessagePrefix {
		return ErrLenLess
	}

	msg.GenBlock 	=  EComm.BytesToHash(bs[ 1 + 256 * 0 : 1 + 256 * 1 ])
	msg.VDBRoot 	=  EComm.BytesToHash(bs[ 1 + 256 * 1 : 1 + 256 * 2 ])
	msg.Header    	=  EComm.BytesToHash(bs[ 1 + 256 * 2 : 1 + 256 * 3 ])
	msg.Header     	=  EComm.BytesToHash(bs[ 1 + 256 * 3 : 1 + 256 * 4 ])

	return nil
}