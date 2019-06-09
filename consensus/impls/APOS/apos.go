package APOS

import (
	"context"
	ACore "github.com/ayachain/go-aya/consensus/core"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	AGroup "github.com/ayachain/go-aya/consensus/core/worker"
	APOSInBlock "github.com/ayachain/go-aya/consensus/impls/APOS/in/block"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ayachain/go-aya/vdb"
	AMsgBlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ipfs/go-ipfs/core"
	"github.com/pkg/errors"
)


var (
	ErrNotSupportMessageTypeExpected = errors.New("not support message type")
)

type APOSConsensusNotary struct {
	ACore.Notary
	mainCVFS vdb.CVFS
	workInd	*core.IpfsNode
	workctx    context.Context
	workCancel context.CancelFunc
	ccmap map[byte]*ACStep.AConsensusStep
}

func NewAPOSConsensusNotary( m vdb.CVFS, ind *core.IpfsNode ) *APOSConsensusNotary {

	ctx, cancel := context.WithCancel(context.Background())

	notary := &APOSConsensusNotary{
		mainCVFS:m,
		workInd:ind,
		workctx:ctx,
		workCancel:cancel,
		ccmap:make(map[byte]*ACStep.AConsensusStep),
	}

	notary.ccmap[AMsgBlock.MessagePrefix] = APOSInBlock.NewConsensusStep(m, ind)

	return notary
}


func (n *APOSConsensusNotary) FireYou() {
	n.workCancel()
}


func (n *APOSConsensusNotary) OnReceiveRawMessage( msg *AKeyStore.ASignedRawMsg ) <- chan ACStep.AConsensusResult {

	replay := make(chan ACStep.AConsensusResult)

	subcc, exist := n.ccmap[msg.Content[0]]
	if !exist {
		replay <- ACStep.AConsensusResult{Err:ErrNotSupportMessageTypeExpected, StepIdentifier:"APOS-Root", Msg:msg}
		return replay
	}

	go func() {

		group := AGroup.NewGroup()

		select {

		case <- n.workctx.Done():
			break

		case ret := <- subcc.DoConsultation( n.workctx, msg, group ):

			if ret.Err == nil {

			}

		}

	}()

	return replay
}