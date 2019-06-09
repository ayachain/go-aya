package workflow

import (
	"bytes"
	"encoding/binary"
	APOSComm "github.com/ayachain/go-aya/consensus/impls/APOS/in/common"
	AWorker "github.com/ayachain/go-aya/consensus/core/worker"
	ARsp "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
)

func DoLockAmount( tx *ATx.Transaction, group *AWorker.TaskBatchGroup, base vdb.CVFS ) error {

	if tx.Verify() {

		if !bytes.Equal(tx.From.Bytes(), tx.To.Bytes()) || string(tx.Data) != "LOCK" {
			return APOSComm.ErrParmasExpected
		}

		//base.Assetses().AvailBalanceMove()
		bindexBs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bindexBs, tx.BlockIndex)

		txHash := tx.GetHash256()
		receiptKey := append(txHash.Bytes(), bindexBs...)

		astfrom, err := base.Assetses().AssetsOf(tx.From.Bytes())
		if err != nil {
			return APOSComm.ErrExpected
		}

		if astfrom.Avail < tx.Value {

			group.Put(
				base.Receipts().DBKey(),
				receiptKey,
				ARsp.RawExpectedResponse(APOSComm.TxReceiptsNotenough),
			)

			return nil

		}

		astfrom.Locked += tx.Value
		astfrom.Avail -= tx.Value

		group.Put( base.Assetses().DBKey(), tx.From.Bytes(), astfrom.Encode() )

		group.Put(
			base.Receipts().DBKey(),
			receiptKey,
			ARsp.RawSusccessResponse(APOSComm.TxConfirm),
		)

		return nil
	}

	return APOSComm.ErrTxVerifyExpected
}


func unlockAmount( tx *ATx.Transaction, group *AWorker.TaskBatchGroup, base vdb.CVFS ) error {

	if tx.Verify() {

		if !bytes.Equal(tx.From.Bytes(), tx.To.Bytes()) || string(tx.Data) != "UNLOCK" {
			return APOSComm.ErrParmasExpected
		}

		//base.Assetses().AvailBalanceMove()
		bindexBs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bindexBs, tx.BlockIndex)

		txHash := tx.GetHash256()
		receiptKey := append(txHash.Bytes(), bindexBs...)

		astfrom, err := base.Assetses().AssetsOf(tx.From.Bytes())
		if err != nil {
			return APOSComm.ErrExpected
		}

		if astfrom.Locked >= tx.Value {

			group.Put(
				base.Receipts().DBKey(),
				receiptKey,
				ARsp.RawExpectedResponse(APOSComm.TxReceiptsNotenough),
			)

			return nil

		}

		astfrom.Locked -= tx.Value
		astfrom.Avail += tx.Value

		group.Put( base.Assetses().DBKey(), tx.From.Bytes(), astfrom.Encode() )

		group.Put(
			base.Receipts().DBKey(),
			receiptKey,
			ARsp.RawSusccessResponse(APOSComm.TxConfirm),
		)

		return nil
	}

	return APOSComm.ErrTxVerifyExpected
}