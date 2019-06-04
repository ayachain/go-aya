package transaction

import (
	"bytes"
	"encoding/json"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Transaction struct {

	AvdbComm.RawDBCoder

	BlockIndex		uint64			`json:"index"`
	From 			EComm.Address	`json:"from"`
	To				EComm.Address	`json:"to"`
	Value			uint64			`json:"value"`
	Children		[]EComm.Hash	`json:"children"`
	Data			[]byte			`json:"data"`
	Steps			uint32			`json:"steps"`
	Price			uint32			`json:"price"`
	Tid				uint64			`json:"tid"`
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
	bs := trsn.Encode()
	return crypto.Keccak256Hash(bs)
}

///// key = Transaction.Bytes + BlockIndex(LittleEndian)
func ( trsn *Transaction ) EncodeRawKey() []byte {

	ibs := AvdbComm.BigEndianBytes( trsn.BlockIndex )
	thash := trsn.GetHash256().Bytes()

	buf := bytes.NewBuffer(thash)
	buf.Write(ibs)

	return buf.Bytes()
}