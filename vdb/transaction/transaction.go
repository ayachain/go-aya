package transaction

import (
	"bytes"
	"encoding/json"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Transaction struct {

	AVdbComm.RawDBCoder

	BlockIndex		uint64			`json:"index"`
	From 			EComm.Address	`json:"from"`
	To				EComm.Address	`json:"to"`
	Value			uint64			`json:"value"`
	Children		[]EComm.Hash	`json:"children"`
	Data			[]byte			`json:"data"`
	Steps			uint32			`json:"steps"`
	Price			uint32			`json:"price"`
	Tid				uint64			`json:"tid"`
	Sig				[]byte			`json:"sig"`
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

///// key = Transaction.Bytes + BlockIndex(LittleEndian)
func ( trsn *Transaction ) EncodeRawKey() []byte {

	ibs := AVdbComm.BigEndianBytes( trsn.BlockIndex )
	thash := trsn.GetHash256().Bytes()

	buf := bytes.NewBuffer(thash)
	buf.Write(ibs)

	return buf.Bytes()
}