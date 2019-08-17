package minined

import (
	"bytes"
	"encoding/json"
	"errors"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/mblock"
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

	MBlock *mblock.MBlock		`json:"MBlock"`
	Batcher cid.Cid				`json:"BatchCID"`
}


func (md *Minined) Encode() []byte {

	bs, err := json.Marshal(md)
	if err != nil {
		return nil
	}
	return bs

}

func (md *Minined) Decode(bs []byte) error {
	return json.Unmarshal(bs, md)
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