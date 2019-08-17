package minerpool

import (
	"context"
	ATxFlow "github.com/ayachain/go-aya/chain/minerpool/tx"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/common"
	ARsp "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"sync"
)

type WorkingStat int

const (
	WorkingStatIdle 	WorkingStat = 0
	WorkingStateBuzy	WorkingStat = 1
)

type Miner interface {

	DoTask( ctx context.Context, task *MiningTask ) *MiningResult

	GetState() WorkingStat

}

type aMiner struct {

	Miner

	state WorkingStat
	smu sync.Mutex
}

func newMiner() Miner {

	return &aMiner{
		state:WorkingStatIdle,
	}
}

func (n *aMiner) GetState() WorkingStat {

	n.smu.Lock()
	defer n.smu.Unlock()

	return n.state
}

func (n *aMiner) DoTask( ctx context.Context, task *MiningTask ) *MiningResult {

	n.smu.Lock()
	defer n.smu.Unlock()

	n.state = WorkingStateBuzy

	defer func() {
		n.smu.Lock()
		n.state = WorkingStatIdle
		n.smu.Unlock()
	}()

	for i, tx := range task.Txs {

		select {
		case <- ctx.Done():

			return &MiningResult{
				Err:ErrContextCancel,
				Task:task,
			}

		default:

		}

		// Is transaction override ?
		txc, err := task.VWriter.Transactions().GetTxCount(tx.From)

		if err != nil {
			continue
		}

		if tx.Tid < txc {
			continue
		}

		// Handle Cost
		if ATxFlow.DoCostHandle( tx, task.VWriter, i ) != nil {
			continue
		}

		// Write tx history
		task.VWriter.Transactions().Put(tx, task.MiningBlock.Index)

		switch tx.Type {

		//case "UNLOCK", "LOCK":
		//	if err := workflow.DoLockAmount(tx, group, vdb); err != nil {
		//		return nil, err
		//	}

		case ATx.NormalTransfer : _ = ATxFlow.DoTransfer( tx, task.VWriter )

		default:
			task.VWriter.Receipts().Put(tx.GetHash256(), task.MiningBlock.Index, ARsp.ExpectedReceipt(APosComm.TxUnsupportTransactionType, nil).Encode())
		}
	}

	batcher := task.VWriter.MergeGroup()
	if batcher == nil {
		return &MiningResult{
			Err:ErrCreateBatch,
			Task:task,
		}
	}

	select {
	case <- ctx.Done():
		return &MiningResult{
			Err:ErrContextCancel,
			Task:task,
		}

	default:

	}

	return &MiningResult{
		Err:nil,
		Batcher:batcher,
		Task:task,
	}

	// should pos block
	//workflow.CanPos( block, cvfs )
}