package tx

import (
	"../../utils"
	"./act"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"strings"
)

type Tx struct {
	Sender			string
	Signature		string
	ActHex			string
}

func NewTx(sender string, acti act.BaseActInf) (tx* Tx) {

	if hex, err := acti.EncodeToHex(); err != nil {
		return nil
	} else {
		tx = &Tx{Sender:sender, Signature:"", ActHex:hex}
	}

	return tx
}

func (tx* Tx) GetActHash() []byte {

	hbs, err := utils.GetHashBytes(tx.ActHex)

	if err != nil {
		return nil
	}

	return hbs
}

func (tx* Tx) VerifySign() (b bool) {

	if len(tx.Signature) <= 0 {
		return false
	}

	hbs, err := utils.GetHashBytes(tx.ActHex)

	if err != nil {
		log.Println(err)
		return false
	}

	hbs = crypto.Keccak256(hbs)

	if err != nil {
		log.Println(err)
		return false
	}

	sigbs, err := hexutil.Decode(tx.Signature)

	if err != nil {
		return false
	}

	if pub, err := crypto.SigToPub(hbs, sigbs); err != nil {
		return false
	} else {
		addr := crypto.PubkeyToAddress(*pub)
		return strings.EqualFold(addr.String(), tx.Sender)
	}

}

func (tx* Tx) EncodeToHex() (hex string, err error) {

	if bs, err := json.Marshal(tx); err == nil {
		return hexutil.Encode(bs),nil
	} else {
		return "", err
	}

}

func (tx* Tx) DecodeFromHex(hex string) error {

	if bs, err := hexutil.Decode(hex); err != nil {
		return nil
	} else {
		return json.Unmarshal(bs, tx)
	}

}

func (tx* Tx) MarshalJson() string {

	if bs, err := json.Marshal(tx); err != nil {
		return ""
	} else {
		return string(bs)
	}

}