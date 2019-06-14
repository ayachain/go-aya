package APOS

import (
	"context"
	ACore "github.com/ayachain/go-aya/consensus/core"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	AGroup "github.com/ayachain/go-aya/consensus/core/worker"
	APOSInBlock "github.com/ayachain/go-aya/consensus/impls/APOS/in/block"
	"github.com/ayachain/go-aya/vdb"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	"github.com/ipfs/go-ipfs/core"
	"github.com/pkg/errors"
)


var (
	ErrNotSupportMessageTypeExpected = errors.New("not support message type")
	ErrNotExistConsensusRule = errors.New("not found rule")
)

type APOSConsensusNotary struct {

	ACore.Notary
	workInd	*core.IpfsNode
	workctx    context.Context
	workCancel context.CancelFunc
	ccmap map[byte]*ACStep.AConsensusStep

}

func NewAPOSConsensusNotary( ind *core.IpfsNode ) *APOSConsensusNotary {

	ctx, cancel := context.WithCancel(context.Background())

	notary := &APOSConsensusNotary{
		workInd:ind,
		workctx:ctx,
		workCancel:cancel,
		ccmap:make(map[byte]*ACStep.AConsensusStep),
	}

	notary.ccmap[AMsgMBlock.MessagePrefix] = APOSInBlock.NewConsensusStep(ind)

	return notary
}


func (n *APOSConsensusNotary) FireYou() {
	n.workCancel()
}


func (n *APOSConsensusNotary) MiningBlock( block *AMsgMBlock.MBlock, cvfs vdb.CacheCVFS ) (*AGroup.TaskBatchGroup, error) {

	subcc, exist := n.ccmap[AMsgMBlock.MessagePrefix]

	if !exist {
		return nil, ErrNotExistConsensusRule
	}

	ctx, cancel := context.WithCancel(n.workctx)

	defer cancel()

	ret := <- subcc.DoConsultation(ctx, block, cvfs)

	if ret.Err != nil {
		return nil, ret.Err
	}

	return cvfs.MergeGroup(), nil
}