package block

import (
	"bytes"
	"encoding/json"
	"errors"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-cid"
)

const MessagePrefix = byte('b')

var (
	ErrMsgPrefix = errors.New("not a chain info message")
)

//var (
//	Genesis = &Block{Index: -4}
//	Curr 	= &Block{Index: -3}
//	Latest 	= &Block{Index: -2}
//	Pending = &Block{Index: -1}
//)

type Block struct {

	AVdbComm.RawDBCoder		`json:"-"`
	AVdbComm.AMessageEncode `json:"-"`

	/// block index
	Index uint64 			`json:"Index"`

	/// chain id
	ChainID string 			`json:"ChainID"`

	/// prev block hash is a ipfs block CID
	Parent  string 			`json:"Parent"`

	/// full block data CID, is cvfs root CID
	ExtraData string 		`json:"ExtraData"`

	/// broadcasting time super master node package this block times.
	Timestamp uint64 		`json:"Timestamp"`

	/// append data
	AppendData []byte 		`json:"Append"`

	/// block sub transactions, is a ipfs block cid
	Txc uint16				`json:"Txc"`
	Txs string				`json:"Txs"`
}

/// only in create a new chain then use
type GenBlock struct {
	Block
	Consensus	string	`json:"Consensus"`
	Award map[string]uint64 `json:"Award"`
}


func ( b *Block ) GetHash() EComm.Hash {

	bs := b.Encode()
	if bs == nil {
		panic("unrecoverable computing exception : Hash")
	}

	return crypto.Keccak256Hash(bs)
}

func ( b *Block ) GetExtraDataCid() cid.Cid {

	c, err := cid.Decode( b.ExtraData )
	if err != nil {
		return cid.Undef
	}

	return c
}

func ( b *Block ) Encode() []byte {

	bs, err := json.Marshal(b)

	if err != nil {
		return nil
	}

	return bs
}

func ( b *Block ) Decode(bs []byte) error {

	//if bs[0] != 'b' {
	//	return errors.New("this raw bytes not a block.")
	//}
	return json.Unmarshal(bs, b)
}

func ( b *Block ) RawMessageEncode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( b.Encode() )

	return buff.Bytes()
}

func ( b *Block ) RawMessageDecode( bs []byte ) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return b.Decode(bs[1:])

}

func ( gb *GenBlock ) GetHash() EComm.Hash {

	bs := gb.Encode()
	if bs == nil {
		panic("unrecoverable computing exception : Hash")
	}

	return crypto.Keccak256Hash(bs)
}

func ( gb *GenBlock ) Encode() []byte {

	bs, err := json.Marshal(gb)

	if err != nil {
		return nil
	}

	return bs
}

func ( gb *GenBlock ) Decode(bs []byte) error {
	return json.Unmarshal(bs, gb)
}

func ( gb *GenBlock ) RawMessageEncode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( gb.Encode() )

	return buff.Bytes()
}

func ( gb *GenBlock ) RawMessageDecode( bs []byte ) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return gb.Decode(bs[1:])

}