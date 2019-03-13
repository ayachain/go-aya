package ChainStruct

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"strings"
)

type DappPerfrom struct {
	DappPath string			`json:"DappPath"`
	Perfrom  string			`json:"Perfrom"`
	Parmas[] string			`json:"Parmas"`
}

type Transaction struct {
	Sender string			`json:"Sender"`
	Value  string			`json:"Value"`
	Data   DappPerfrom		`json:"Data"`
	Sign   string			`json:"Sign"`
}

func (t* Transaction) GetContentHash() (b[] byte, err error){

	c, e := json.Marshal(t.Data)

	if e != nil {
		return nil, e
	}

	cBuffer := bytes.NewBufferString(t.Sender + t.Value)
	cBuffer.Write(c)

	return crypto.Keccak256(cBuffer.Bytes()), nil
}

func (t* Transaction) Verify() (v bool, err error) {

	hash, err := t.GetContentHash()

	if err != nil {
		return false, err
	}

	signByte,err := hexutil.Decode(t.Sign)

	if err != nil {
		return false, err
	}

	pub, err := secp256k1.RecoverPubkey(hash, signByte)

	if err != nil {
		return false,err
	}

	pubkey,err := crypto.UnmarshalPubkey(pub)

	if err != nil {
		return false,err
	}

	return strings.EqualFold(crypto.PubkeyToAddress(*pubkey).Hex(), t.Sender), nil

}

func GenTransaction(sender string, value string, data DappPerfrom, key* ecdsa.PrivateKey) (r* Transaction, err error) {

	r = &Transaction{sender, value, data, ""}

	hash, err := r.GetContentHash()

	if err != nil {
		return nil, err
	}

	signature, err := crypto.Sign(hash, key)

	if err != nil {
		return nil, err
	}

	r.Sign = hexutil.Encode(signature)

	return r, nil

}

func (t* Transaction) EncodeToHex() (hex string, err error){

	bs, err := json.Marshal(t)

	if err != nil {
		return "", err
	}

	return hexutil.Encode(bs), nil
}

func (t* Transaction) DecodeFromHex( hex string ) (err error) {

	bs, err := hexutil.Decode(hex)

	if err != nil {
		return err
	}

	return json.Unmarshal(bs, t)
}