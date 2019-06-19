package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const MessagePrefix = byte('t')

var (
	ErrMsgPrefix = errors.New("not a chain info message")
)

type Transaction struct {

	AVdbComm.RawDBCoder				`json:"-"`
	AVdbComm.AMessageEncode			`json:"-"`

	BlockIndex		uint64			`json:"Index"`
	From 			EComm.Address	`json:"From"`
	To				EComm.Address	`json:"To"`
	Value			uint64			`json:"Value"`
	Children		[]EComm.Hash	`json:"Children"`
	Data			[]byte			`json:"Data"`
	Steps			uint32			`json:"Steps"`
	Price			uint32			`json:"Price"`
	Tid				uint64			`json:"Tid"`
	Sig				[]byte			`json:"Sig"`

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

	for _, hs := range trsn.Children {
		buff.Write(hs.Bytes())
	}
	buff.Write(trsn.Data)
	buff.Write( AVdbComm.BigEndianBytesUint32(trsn.Steps) )
	buff.Write( AVdbComm.BigEndianBytesUint32(trsn.Price) )
	buff.Write( AVdbComm.BigEndianBytes(trsn.Tid) )
	//buff.Write( trsn.Sig )

	return crypto.Keccak256Hash(buff.Bytes())
}

func ( trsn *Transaction ) Verify() bool {

	hs := trsn.GetHash256()

	pubkey, err := crypto.Ecrecover( hs.Bytes(), trsn.Sig )
	if err != nil {
		return false
	}

	return bytes.Equal(pubkey, trsn.From.Bytes())

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