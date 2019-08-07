package electoral

import (
	"bytes"
	"encoding/json"
	"errors"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
)

const MessagePrefix = byte('v')

var (
	ErrMsgPrefix    = errors.New("not a electoral info message")
)

type Electoral struct {

	AvdbComm.RawDBCoder				`json:"-"`
	AvdbComm.AMessageEncode			`json:"-"`

	BestIndex		uint64			`json:"BestIndex"`
	BlockIndex		uint64			`json:"PackIndex"`
	From 			EComm.Address	`json:"From,omitempty"`
	ToPeerId		string			`json:"ToPeerId,omitempty"`
	Time			int64			`json:"Time"`
}

func (vote *Electoral) Encode() []byte {

	bs, err := json.Marshal(vote)
	if err != nil {
		return nil
	}

	return bs
}

func (vote *Electoral) Decode(bs []byte) error {

	return json.Unmarshal(bs, vote)
}


func (vote *Electoral) RawMessageEncode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( vote.Encode() )

	return buff.Bytes()

}

func (vote *Electoral) RawMessageDecode( bs []byte ) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return vote.Decode(bs[1:])

}


type ATxPackerState string

const (

	ATxPackStateUnknown			ATxPackerState	=	"Unknown"
	ATxPackStateMaster 			ATxPackerState	=	"Master"
	ATxPackStateNextMaster		ATxPackerState	=	"NextMaster"
	ATxPackStateFollower		ATxPackerState	=	"Follower"
	ATxPackStateLookup			ATxPackerState  =   "Looking"
)