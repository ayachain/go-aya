package APOS

import (
	"context"
	ACore "github.com/ayachain/go-aya/consensus/core"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	APosSign "github.com/ayachain/go-aya/consensus/impls/APOS/signaturer"
	APosDog "github.com/ayachain/go-aya/consensus/impls/APOS/watchdog"
	AMsgBlock "github.com/ayachain/go-aya/chain/message/block"
	APosWorker "github.com/ayachain/go-aya/consensus/impls/APOS/worker"
	APosExecutor "github.com/ayachain/go-aya/consensus/impls/APOS/executor"
	APosDataLoad "github.com/ayachain/go-aya/consensus/impls/APOS/dataloader"
	"github.com/ayachain/go-aya/vdb"
	"github.com/ipfs/go-ipfs/core"
	"github.com/libp2p/go-libp2p-pubsub"
)

type APOSConsensusNotary struct {
	ACore.Notary
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
	dog.SetRule( AMsgBlock.MessagePrefix, blockRule )

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

func (n *APOSConsensusNotary) OnReceiveMessage( msg *pubsub.Message ) error {
	return n.HiDog().TakeMessage(msg)
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
