package signaturer

import (
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
)

func (s *Signturer) Identifier( ) string {
	return s.identifier
}

func (s *Signturer) SetNextStep( ns ACStep.ConsensusStep ) {
	s.nextStep = ns
}

func (s *Signturer) ChannelAccept() chan *ADog.MsgFromDogs {
	return s.acceptChan
}

func (s *Signturer) NextStep() ACStep.ConsensusStep {
	return s.nextStep
}

func (s *Signturer) Consensued( *ADog.MsgFromDogs ) {
	panic("nonreversible consensus expected")
}