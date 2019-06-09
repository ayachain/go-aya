package keystore

import (
	AMsgTransation "github.com/ayachain/go-aya/vdb/transaction"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	EKeyStore "github.com/ethereum/go-ethereum/accounts/keystore"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"strings"
)

var _akeystoreInstance *EKeyStore.KeyStore = nil

func Init( ksdir string ) {
	if _akeystoreInstance == nil {

		_akeystoreInstance = EKeyStore.NewKeyStore( ksdir, EKeyStore.LightScryptN, EKeyStore.LightScryptP)

		if _akeystoreInstance == nil {
			panic( "Init KeyStore failed." )
		}
	}
}


func ShareInstance() *EKeyStore.KeyStore {
	return _akeystoreInstance
}

func FindAccount( hexAddr string ) (EAccount.Account, error) {

	aks := ShareInstance()
	if aks == nil {
		return EAccount.Account{}, errors.New("keystore services not daemon")
	}

	for _, acc := range aks.Accounts() {
		if strings.EqualFold( acc.Address.String(), hexAddr ) {
			return acc, nil
		}
	}

	return EAccount.Account{}, errors.New("not found")
}

func signHash( hash EComm.Hash, acc EAccount.Account ) ([]byte, error) {
	return _akeystoreInstance.SignHash(acc, hash.Bytes())
}

func SignTransaction( tx *AMsgTransation.Transaction, acc EAccount.Account ) error {

	hash := tx.GetHash256()

	sig, err := signHash(hash, acc)
	if err != nil {
		return err
	}

	tx.Sig = sig

	return nil
}