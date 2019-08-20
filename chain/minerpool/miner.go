package minerpool

import (
	ATxFlow "github.com/ayachain/go-aya/chain/minerpool/tx"
	ACComm "github.com/ayachain/go-aya/chain/common"
	ARsp "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
)

func (mp *aMinerPool) doTask( task *MiningTask ) *MiningResult {

	for i, tx := range task.Txs {

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
			task.VWriter.Receipts().Put(tx.GetHash256(), task.MiningBlock.Index, ARsp.ExpectedReceipt(ACComm.TxUnsupportTransactionType, nil).Encode())
		}
	}

	batcher := task.VWriter.MergeGroup()
	if batcher == nil {
		return &MiningResult{
			Err:ErrCreateBatch,
			Task:task,
		}
	}

	return &MiningResult{
		Err:nil,
		Batcher:batcher,
		Task:task,
	}

	// should pos block
	//workflow.CanPos( block, cvfs )
}