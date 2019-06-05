package APOS

import (
	"context"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	APosDataLoad "github.com/ayachain/go-aya/consensus/impls/APOS/dataloader"
	APosExecutor "github.com/ayachain/go-aya/consensus/impls/APOS/executor"
	APosSign "github.com/ayachain/go-aya/consensus/impls/APOS/signaturer"
	APosDog "github.com/ayachain/go-aya/consensus/impls/APOS/watchdog"
	APosWorker "github.com/ayachain/go-aya/consensus/impls/APOS/worker"
	"github.com/ayachain/go-aya/vdb"
	"github.com/ipfs/go-ipfs/core"
)

type APOSConsensusNotary struct {

	mainCVFS vdb.CVFS
	mydog    *APosDog.Dog

	workctx    context.Context
	workCancel context.CancelFunc

	cclist []*ACStep.ConsensusChain

}


func NewAPOSConsensusNotary( m vdb.CVFS, ind *core.IpfsNode ) *APOSConsensusNotary {

	dog := APosDog.NewDog()

	var list []*ACStep.ConsensusChain

	// config block message consensus rule
	blockRule := ACStep.NewConsensusChain("APOS")
	list = append(list, blockRule)

	blockRule.AppendSteps(
		APosSign.NewSignturer(),
		APosDataLoad.NewDataLoader(ind),
		APosWorker.NewWorker( ind, m ),
		APosExecutor.NewExecutor(m),
		)

	blockRule.LinkAllStep()
	dog.SetRule('b', blockRule )


	ctx,cancel := context.WithCancel(context.Background())

	return &APOSConsensusNotary{
		mainCVFS:m,
		mydog:dog,
		cclist:list,
		workctx:ctx,
		workCancel:cancel,
	}

}


func (n *APOSConsensusNotary) HiDog() *APosDog.Dog {
	return n.mydog
}


func (n *APOSConsensusNotary) FireYou() {
	n.workCancel()
}


func (n *APOSConsensusNotary) StartWorking() {

	for _, cc := range n.cclist {

		root := cc.GetStepRoot()

		root.StartListenAccept(n.workctx)

		for root.NextStep() != nil {

			root = root.NextStep()

			root.StartListenAccept(n.workctx)

		}
	}

}