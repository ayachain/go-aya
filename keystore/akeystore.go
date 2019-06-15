package keystore

import (
	AMsgTransation "github.com/ayachain/go-aya/vdb/transaction"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	EKeyStore "github.com/ethereum/go-ethereum/accounts/keystore"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/core"
	"github.com/pkg/errors"
	"strings"
)


const CoinBaseKey = "aya.wallet.coinbase"

var _akeystoreInstance *EKeyStore.KeyStore = nil

var _coinBaseAccount EAccount.Account

var _ind *core.IpfsNode

func Init( ksdir string, ind *core.IpfsNode ) {

	if _akeystoreInstance == nil {

		_ind = ind

		_akeystoreInstance = EKeyStore.NewKeyStore( ksdir, EKeyStore.LightScryptN, EKeyStore.LightScryptP)

		if _akeystoreInstance == nil {
			panic( "Init KeyStore failed." )
		}

		dsk := datastore.NewKey(CoinBaseKey)
		exist, err := ind.Repo.Datastore().Has(dsk)

		if err == nil || exist {

			addrbs, err := ind.Repo.Datastore().Get(dsk)
			if err != nil {
				return
			}

			addr := EComm.BytesToAddress(addrbs)

			for _, acc := range _akeystoreInstance.Accounts() {

				if acc.Address.Hash() == addr.Hash() {

					_coinBaseAccount = acc

					return
				}
			}

		}

	}

}

func ShareInstance() *EKeyStore.KeyStore {
	return _akeystoreInstance
}

func SetCoinBaseAddress( account EAccount.Account ) error {

	_coinBaseAccount = account

	dsk := datastore.NewKey(CoinBaseKey)

	if err := _ind.Repo.Datastore().Put(dsk, account.Address.Bytes()); err != nil {
		return err
	}

	return nil
}

func GetCoinBaseAddress() EAccount.Account {
	return _coinBaseAccount
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

func SignTransaction( tx *AMsgTransation.Transaction, acc EAccount.Account ) error {

	hash := tx.GetHash256()

	sig, err := signHash(hash, acc)
	if err != nil {
		return err
	}

	tx.Sig = sig

	return nil
}

func signHash( hash EComm.Hash, acc EAccount.Account ) ([]byte, error) {
	return _akeystoreInstance.SignHash(acc, hash.Bytes())
}
