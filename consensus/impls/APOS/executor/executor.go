package executor

import (
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/common"
	"github.com/ayachain/go-aya/vdb"
)

type Executor struct {

	ACStep.ConsensusStep

	nextStep ACStep.ConsensusStep

	acceptChan chan *ADog.MsgFromDogs

	identifier string

	cvfs vdb.CVFS
}

func NewExecutor( vfs vdb.CVFS ) *Executor {
	return &Executor{
		cvfs:vfs,
		identifier : "APOS-Step-4-Executor",
		acceptChan : make(chan *ADog.MsgFromDogs, APosComm.StepExecutorChanSize),
	}
}