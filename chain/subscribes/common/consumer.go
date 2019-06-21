package common

import (
	AKeyStore "github.com/ayachain/go-aya/keystore"
)

type Consumer interface {

	DoConsume( msg *AKeyStore.ASignedRawMsg ) <- chan struct{}

}