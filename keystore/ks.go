package keystore

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	"github.com/ethereum/go-ethereum/crypto"
)

type Keystore struct {
	pkey* 	ecdsa.PrivateKey
}

func (ks *Keystore) GenerateKey() error {
	if key, err := crypto.GenerateKey(); err != nil {
		return err
	} else {
		ks.pkey = key
		return nil
	}
}

func (ks *Keystore) PrivateKey() string {
	return hex.EncodeToString(ks.pkey.D.Bytes())
}

func (ks *Keystore) Address() string {
	return crypto.PubkeyToAddress(ks.pkey.PublicKey).Hex()
}

func (ks *Keystore) SignTx(tx *Atx.Tx) error {

	if len(tx.ActHex) <= 0 {
		return errors.New("Keysotre: Tx is empty, can't sign.")
	}

	sig, err := crypto.Sign(crypto.Keccak256(tx.GetActHash()), ks.pkey)

	if err != nil {
		panic(err)
	}

	tx.Sender = ks.Address()
	tx.Signature = "0x" + hex.EncodeToString(sig)

	return nil
}

func (ks *Keystore) CreateSignedTx(act Act.BaseActInf) *Atx.Tx {

	stx := Atx.NewTx(ks.Address(), act)

	if err := ks.SignTx(stx); err != nil {
		return nil
	} else {
		return stx
	}

}

var peerDefaultKS = &Keystore{}

func DefaultPeerKS() *Keystore {

	if peerDefaultKS.pkey == nil {
		if err := peerDefaultKS.GenerateKey(); err != nil {
			return nil
		}
	}

	return peerDefaultKS
}