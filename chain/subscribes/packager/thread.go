package packager

import (
	"context"
	AChainComm "github.com/ayachain/go-aya/chain/subscribes/common"
	AKeyStore "github.com/ayachain/go-aya/keystore"
)


type APackagerThread struct {
	AChainComm.AThread
}


func (consumer *APackagerThread) DoConsume( msg *AKeyStore.ASignedRawMsg ) <- chan struct{} {

	cc := make(chan struct{})

	cc <- nil

	return cc
}


func (consumer *APackagerThread) DoProduce(ctx context.Context) (*AKeyStore.ASignedRawMsg, error) {
	return &AKeyStore.ASignedRawMsg{}, nil
}