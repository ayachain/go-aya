package step

import (
	"context"
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	Avdb "github.com/ayachain/go-aya/vdb"
	"github.com/ipfs/go-ipfs/core"
)

type AConsensusFunc func(interface{}, Avdb.CVFS, *core.IpfsNode, *AWork.TaskBatchGroup ) (interface{}, error)


type AConsensusResult struct {
	Err error
	StepIdentifier string
	Msg interface{}
}


type AConsensusStep struct {

	next *AConsensusStep
	identifier string
	cfun AConsensusFunc

	cvfs Avdb.CVFS
	ind *core.IpfsNode
	working bool
}


func NewStepRoot( identifier string, cvfs Avdb.CVFS, ind *core.IpfsNode, fun AConsensusFunc ) *AConsensusStep {

	return &AConsensusStep{
		next:nil,
		identifier:identifier,
		cfun:fun,
		cvfs:cvfs,
		ind:ind,
	}

}

func (parent *AConsensusStep) LinkNext ( identifier string, fun AConsensusFunc ) *AConsensusStep {

	sub := &AConsensusStep{
		next:nil,
		identifier:identifier,
		cfun:fun,
		cvfs:parent.cvfs,
		ind:parent.ind,
	}

	parent.next = sub

	return sub
}

func (bl *AConsensusStep) DoConsultation( ctx context.Context, msg interface{}, group *AWork.TaskBatchGroup ) <- chan AConsensusResult {

	cc := make(chan AConsensusResult, 1)

	go func( stp *AConsensusStep ) {

		if rmsg, err := bl.cfun( msg, bl.cvfs, bl.ind, group ); err != nil {
			cc <- AConsensusResult{ err, stp.identifier, msg }
		} else {
			if bl.next != nil {
				cc <- <- bl.next.DoConsultation( ctx, rmsg, group )
			} else {
				cc <- AConsensusResult{ nil, stp.identifier, msg }
			}
		}

	}(bl)

	return cc
}


func (bl *AConsensusStep) Identifier( ) string {
	return bl.identifier
}