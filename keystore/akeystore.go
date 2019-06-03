package keystore

import (
	EKeyStore "github.com/ethereum/go-ethereum/accounts/keystore"
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

func ShareInstance(  ) *EKeyStore.KeyStore {
	return _akeystoreInstance
}