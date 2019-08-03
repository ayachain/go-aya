package chaininfo

import (
	"bytes"
	"encoding/json"
	"errors"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AvdbComm"github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
)

const MessagePrefix = byte('i')

var (
	ErrMsgPrefix    = errors.New("not a chain info message")
)

type ChainInfo struct {

	AvdbComm.RawDBCoder				`json:"-"`
	AvdbComm.AMessageEncode			`json:"-"`

	GenHash 		EComm.Hash	 	`json:"GenHash, omitempty"`
	VDBRoot			cid.Cid			`json:"VDBRoot, omitempty"`
	Indexes			cid.Cid 		`json:"Indexes, omitempty"`
	LatestBlock		*ABlock.Block	`json:"LatestBlock, omitempty"`
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