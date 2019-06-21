package common

import (
	"context"
	AKeyStore "github.com/ayachain/go-aya/keystore"
)

type Producer interface {
	DoProduce(ctx context.Context) (*AKeyStore.ASignedRawMsg, error)
}