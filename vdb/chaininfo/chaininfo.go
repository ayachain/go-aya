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

type ChainInfo struct {
	AvdbComm.RawDBCoder			`json:"-"`
	AvdbComm.AMessageEncode		`json:"-"`

	GenBlock 	EComm.Hash	 	`json:"G"`
	VDBRoot		EComm.Hash		`json:"V"`
	Header		EComm.Hash 		`json:"H"`
	Block		EComm.Hash		`json:"B"`
}

func (info *ChainInfo) Encode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( info.GenBlock.Bytes() )
	buff.Write( info.VDBRoot.Bytes() )
	buff.Write( info.Header.Bytes() )
	buff.Write( info.Block.Bytes() )

	return buff.Bytes()
}

func (info *ChainInfo) Decode(bs []byte) error {

	if len(bs) != MessageSize {
		return ErrMsgPrefix
	}

	if bs[0] != MessagePrefix {
		return ErrLenLess
	}

	info.GenBlock 	=  EComm.BytesToHash(bs[ 1 + 32 * 0 : 1 + 32 * 1 ])
	info.VDBRoot 	=  EComm.BytesToHash(bs[ 1 + 32 * 1 : 1 + 32 * 2 ])
	info.Header    	=  EComm.BytesToHash(bs[ 1 + 32 * 2 : 1 + 32 * 3 ])
	info.Header     =  EComm.BytesToHash(bs[ 1 + 32 * 3 : 1 + 32 * 4 ])

	return nil
}

func (info *ChainInfo) RawMessageEncode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( info.Encode() )

	return buff.Bytes()

}

func (info *ChainInfo) RawMessageDecode( bs []byte ) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return info.Decode(bs[1:])


}