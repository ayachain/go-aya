package minined

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


type Minined struct {
	AvdbComm.RawDBCoder			`json:"-"`
	AvdbComm.AMessageEncode		`json:"-"`
	
	MBlockHash EComm.Hash		`json:"MiningBlockHash, omitempty"`
	RetCID cid.Cid				`json:"BatchGroupCID, omitempty"`
}


func (md *Minined) Encode() []byte {

	buff := bytes.NewBuffer([]byte{})
	buff.Write( md.MBlockHash.Bytes() )
	buff.Write( md.RetCID.Bytes() )

	return buff.Bytes()
}

func (md *Minined) Decode(bs []byte) error {

	var err error

	md.MBlockHash = EComm.BytesToHash(bs[0 : 32])
	md.RetCID, err = cid.Cast(bs[32:])

	if err != nil {
		return err
	}

	return nil
}


func (md *Minined) RawMessageEncode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( md.Encode() )

	return buff.Bytes()

}

func (md *Minined) RawMessageDecode( bs []byte ) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return md.Decode(bs[1:])
}