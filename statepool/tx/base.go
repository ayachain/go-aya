package tx

import (
	"../../utils"
	"./act"
	"bytes"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type Tx struct {

	Sender			string
	Signature		string
	Action			*act.BaseAct

}

func NewTx(sender string, act* act.BaseAct) (tx* Tx) {

	tx = &Tx{Sender:sender, Signature:nil, Action:act}

	return tx
}

func (tx* Tx) VerifySign() (b bool) {

	if len(tx.Signature) <= 0 {
		return false
	}

	hbs, err := utils.GetHashBytes(tx.Action)

	if err != nil {
		return false
	}

	if pub, err := secp256k1.RecoverPubkey(hbs, []byte(tx.Signature)); err != nil {
		return false
	} else {
		return bytes.Equal(pub, []byte(tx.Sender))
	}

}