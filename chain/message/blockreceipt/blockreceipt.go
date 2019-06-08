package blockreceipt

import (
	"bytes"
	"errors"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
)

var (
	ErrMsgPrefix = errors.New("not a chain info message")
)

// 1 byte prefix + 4 * Hash(32Byte) = 129 byte
const MessagePrefix = byte('r')


type MsgRawMiningReceipt struct {
	AvdbComm.RawDBCoder
	MBlockHash EComm.Hash
	RetCID cid.Cid
}


func (msg *MsgRawMiningReceipt) Encode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( msg.MBlockHash.Bytes() )

	buff.Write( msg.RetCID.Bytes() )

	return buff.Bytes()
}


func (msg *MsgRawMiningReceipt) Decode(bs []byte) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	var err error
	msg.MBlockHash = EComm.BytesToHash(bs[1 : 257])
	msg.RetCID, err = cid.Cast(bs[258:])
	if err != nil {
		return err
	}

	return nil
}
