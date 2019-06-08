package tx

import (
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/common"
)

type TxSender struct {

	ACStep.ConsensusStep

	nextStep ACStep.ConsensusStep

	acceptChan chan *ADog.MsgFromDogs

	identifier string

}

func NewTxSender() ACStep.ConsensusStep {

	stp := &TxSender{
		identifier : "TxSender",
		acceptChan : make(chan *ADog.MsgFromDogs, APosComm.StepTxSenderChanSize),
	}

	return stp
}