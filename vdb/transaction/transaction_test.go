package transaction

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
)


func TestTransaction_Coder(t *testing.T) {

	tx := &Transaction{
		BlockIndex:1,
		From:common.HexToAddress("0x7353e40668d1f5714dc7265e71cfdc345eae90ed"),
		To:common.HexToAddress("0x47fe409369ea9eaed40564ed6092b34ffe9c365e"),
		Value:1000,
		Children: []common.Hash{
			crypto.Keccak256Hash([]byte("SubTx_1")),
			crypto.Keccak256Hash([]byte("SubTx_2")),
			crypto.Keccak256Hash([]byte("SubTx_3")),
			crypto.Keccak256Hash([]byte("SubTx_4")),
			crypto.Keccak256Hash([]byte("SubTx_5")),
		},
		Steps:65546,
		Price:1000,
		Data:[]byte(""),
	}

	encodebs := tx.Encode()

	if len(encodebs) <= 0 {
		t.Fatal("Encode expected")
	}

	dcdtx := new(Transaction)

	if err := dcdtx.Decode(encodebs); err != nil {
		t.Fatal(err)
	}

	if dcdtx.From.String() != tx.From.String() {
		t.Fatal( "Decode after encode from address is not equal" )
	}

	if dcdtx.To.String() != tx.To.String() {
		t.Fatal( "Decode after encode to address is not equal" )
	}
}
