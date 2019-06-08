package workflow

import (
	"bytes"
	"encoding/binary"
	AWorker "github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ayachain/go-aya/consensus/impls/APOS/in/block"
	ARsp "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
)


func DoLockAmount( tx *ATx.Transaction, group *AWorker.TaskBatchGroup, base vdb.CVFS ) error {

	if tx.Verify() {

		if !bytes.Equal(tx.From.Bytes(), tx.To.Bytes()) || string(tx.Data) != "LOCK" {
			return block.ErrParmasExpected
		}

		//base.Assetses().AvailBalanceMove()
		bindexBs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bindexBs, tx.BlockIndex)

		txHash := tx.GetHash256()
		receiptKey := append(txHash.Bytes(), bindexBs...)

		astfrom, err := base.Assetses().AssetsOf(tx.From.Bytes())
		if err != nil {
			return block.ErrExpected
		}

		if astfrom.Avail < tx.Value {

			group.Put(
				base.Receipts().DBKey(),
				receiptKey,
				ARsp.RawExpectedResponse(block.TxReceiptsNotenough),
			)

			return nil

		}

		astfrom.Locked += tx.Value
		astfrom.Avail -= tx.Value

		group.Put( base.Assetses().DBKey(), tx.From.Bytes(), astfrom.Encode() )

		group.Put(
			base.Receipts().DBKey(),
			receiptKey,
			ARsp.RawSusccessResponse(block.TxConfirm),
		)

		return nil
	}

	return block.ErrTxVerifyExpected
}


func unlockAmount( tx *ATx.Transaction, group *AWorker.TaskBatchGroup, base vdb.CVFS ) error {

	if tx.Verify() {

		if !bytes.Equal(tx.From.Bytes(), tx.To.Bytes()) || string(tx.Data) != "UNLOCK" {
			return block.ErrParmasExpected
		}

		//base.Assetses().AvailBalanceMove()
		bindexBs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bindexBs, tx.BlockIndex)

		txHash := tx.GetHash256()
		receiptKey := append(txHash.Bytes(), bindexBs...)

		astfrom, err := base.Assetses().AssetsOf(tx.From.Bytes())
		if err != nil {
			return block.ErrExpected
		}

		if astfrom.Locked >= tx.Value {

			group.Put(
				base.Receipts().DBKey(),
				receiptKey,
				ARsp.RawExpectedResponse(block.TxReceiptsNotenough),
			)

			return nil

		}

		astfrom.Locked -= tx.Value
		astfrom.Avail += tx.Value

		group.Put( base.Assetses().DBKey(), tx.From.Bytes(), astfrom.Encode() )

		group.Put(
			base.Receipts().DBKey(),
			receiptKey,
			ARsp.RawSusccessResponse(block.TxConfirm),
		)

		return nil
	}

	return block.ErrTxVerifyExpected
}