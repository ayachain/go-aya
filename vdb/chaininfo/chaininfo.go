package chaininfo

import (
	"bytes"
	"encoding/json"
	"errors"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-cid"
)

const MessagePrefix = byte('i')

var (
	ErrMsgPrefix    = errors.New("not a chain info message")
)

type ChainInfo struct {

	AvdbComm.RawDBCoder				`json:"-"`
	AvdbComm.AMessageEncode			`json:"-"`

	ChainID 		string			`json:"ChainID"`
	Indexes			cid.Cid 		`json:"Indexes,omitempty"`
	BlockIndex		uint64			`json:"BlockIndex"`
	FinalCVFS		cid.Cid			`json:"Final"`
}

func (info *ChainInfo) Encode() []byte {

	bs, err := json.Marshal(info)
	if err != nil {
		return nil
	}

	return bs
}

func (info *ChainInfo) Decode(bs []byte) error {

	return json.Unmarshal(bs, info)
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