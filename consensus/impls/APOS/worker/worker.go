package worker

import (
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	APos "github.com/ayachain/go-aya/consensus/impls/APOS"
	"github.com/ayachain/go-aya/vdb"
	"github.com/ipfs/go-ipfs/core"
)

type Worker struct {

	nextStep ACStep.ConsensusStep

	acceptChan chan *ADog.MsgFromDogs

	identifier string

	ind *core.IpfsNode

	cvfs *vdb.CVFS

}


func NewWorker(ind *core.IpfsNode, cvfs vdb.CVFS) *Worker {

	return &Worker{
		identifier : "APOS-Step-3-Worker",
		acceptChan : make( chan *ADog.MsgFromDogs, APos.StepWorkerChanSize ),
	}

}