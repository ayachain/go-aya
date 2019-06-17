package APOS

import (
	"context"
	"encoding/json"
	ACore "github.com/ayachain/go-aya/consensus/core"
	AGroup "github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ayachain/go-aya/consensus/impls/APOS/workflow"
	"github.com/ayachain/go-aya/vdb"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/pkg/errors"
)


var (
	ErrNotSupportMessageTypeExpected = errors.New("not support message type")
	ErrNotExistConsensusRule = errors.New("not found rule")
)

type APOSConsensusNotary struct {

	ACore.Notary

	workctx    context.Context

	workCancel context.CancelFunc

	ind *core.IpfsNode
}

func NewAPOSConsensusNotary( ind *core.IpfsNode ) *APOSConsensusNotary {

	ctx, cancel := context.WithCancel(context.Background())

	notary := &APOSConsensusNotary{
		workctx:ctx,
		workCancel:cancel,
		ind:ind,
	}

	return notary
}

func (n *APOSConsensusNotary) MiningBlock( block *AMsgMBlock.MBlock, cvfs vdb.CacheCVFS ) (*AGroup.TaskBatchGroup, error) {

	txsCid, err := cid.Decode(block.Txs)
	if err != nil {
		return nil, err
	}

	iblock, err := n.ind.Blocks.GetBlock(context.TODO(), txsCid)
	if err != nil {
		return nil, err
	}

	txlist := make([]*ATx.Transaction, block.Txc)
	if err := json.Unmarshal(iblock.RawData(), &txlist); err != nil {
		return nil, err
	}

	for _, tx := range txlist {

		txc, err := cvfs.Transactions().GetTxCount(tx.From)
		if err != nil {
			continue
		}

		if txc != tx.Tid - 1 {
			continue
		}

		cvfs.Transactions().Put(tx, block.Index)

		switch string(tx.Data) {

		//case "UNLOCK", "LOCK":
		//	if err := workflow.DoLockAmount(tx, group, vdb); err != nil {
		//		return nil, err
		//	}

		default:

			if err := workflow.DoTransfer(tx, cvfs); err != nil {
				return nil, err
			}

		}

	}

	return cvfs.MergeGroup(), nil
}