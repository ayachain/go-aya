package signaturer

import (
	ASign "github.com/ayachain/go-aya/consensus/core/signaturer"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
)

type Signturer struct {

	ASign.SignatureAPI

	nextStep ACStep.ConsensusStep

	acceptChan chan *ADog.MsgFromDogs

	identifier string

}

func NewSignturer() *Signturer {
	return &Signturer{
		identifier : "APOS-Step-1-Signature",
	}
}