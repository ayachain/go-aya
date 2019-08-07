package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

const MessagePrefix = byte('t')

var (
	ErrMsgPrefix = errors.New("not a chain info message")
)

type TransactionType int

const (

	NormalTransfer		TransactionType = 0

	UnrecordedTransfer 	TransactionType = 1
)


var StaticCostMapping = map[TransactionType]uint64{

	NormalTransfer 		: 250000,

	UnrecordedTransfer 	: 100000,

}


type Transaction struct {

	AVdbComm.RawDBCoder				`json:"-"`
	AVdbComm.AMessageEncode			`json:"-"`

	BlockIndex		uint64			`json:"Index"`
	From 			EComm.Address	`json:"From,omitempty"`
	To				EComm.Address	`json:"To,omitempty"`
	Value			uint64			`json:"Value"`
	Data			[]byte			`json:"Data,omitempty"`
	Type			TransactionType	`json:"Type"`
	Tid				uint64			`json:"Tid"`
	Sig				[]byte			`json:"Sig,omitempty"`

}

func ( trsn *Transaction ) Encode() []byte {

	bs, err := json.Marshal(trsn)
	if err != nil {
		return nil
	}

	return bs
}

func ( trsn *Transaction ) Decode( bs []byte ) error {
	return json.Unmarshal(bs, trsn)
}

func ( trsn *Transaction ) GetHash256( ) EComm.Hash {

	buff := bytes.NewBuffer([]byte("AyaTransactionPrefix"))

	buff.Write( AVdbComm.BigEndianBytes(trsn.BlockIndex) )
	buff.Write( trsn.From.Bytes() )
	buff.Write( trsn.To.Bytes() )
	buff.Write( AVdbComm.BigEndianBytes(trsn.Value) )
	buff.Write( trsn.Data )
	buff.Write( AVdbComm.BigEndianBytes(trsn.Tid) )
	//buff.Write( trsn.Sig )

	return crypto.Keccak256Hash(buff.Bytes())
}

func ( trsn *Transaction ) Verify() bool {

	hs := trsn.GetHash256()

	pubkey, err := crypto.SigToPub(hs.Bytes(), trsn.Sig)
	if err != nil {
		return false
	}

	from := crypto.PubkeyToAddress(*pubkey)

	return strings.EqualFold(from.String(), trsn.From.String())
}

func ( trsn *Transaction ) EncodeRawKey() []byte {

	ibs := AVdbComm.BigEndianBytes( trsn.BlockIndex )
	thash := trsn.GetHash256().Bytes()

	buf := bytes.NewBuffer(thash)
	buf.Write(ibs)

	return buf.Bytes()
}

func ( trsn *Transaction ) RawMessageEncode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( trsn.Encode() )

	return buff.Bytes()
}

func ( trsn *Transaction ) RawMessageDecode( bs []byte ) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return trsn.Decode(bs[1:])

}