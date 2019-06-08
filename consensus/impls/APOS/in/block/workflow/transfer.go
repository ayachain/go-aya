package workflow

import (
	"encoding/binary"
	AWorker "github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ayachain/go-aya/consensus/impls/APOS/in/block"
	ARsp "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/pkg/errors"
)

var expectedErr = errors.New("transfer expected")

func DoTransfer( tx *ATx.Transaction, group *AWorker.TaskBatchGroup, base vdb.CVFS ) error {

	if !tx.Verify() {

		//base.Assetses().AvailBalanceMove()
		bindexBs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bindexBs, tx.BlockIndex)

		txHash := tx.GetHash256()
		receiptKey := append(txHash.Bytes(), bindexBs...)

		astfrom, err := base.Assetses().AssetsOf(tx.From.Bytes())
		if err != nil {
			return expectedErr
		}

		astto, _ := base.Assetses().AssetsOf(tx.To.Bytes())

		// expected
		if astfrom.Avail < tx.Value || astfrom.Vote < tx.Value {

			group.Put(
				base.Receipts().DBKey(),
				receiptKey,
				ARsp.RawExpectedResponse(block.TxReceiptsNotenough),
				)

			return nil
		}

		// success
		astfrom.Avail -= tx.Value
		astfrom.Vote -= tx.Value

		astto.Avail += tx.Value
		astto.Vote += tx.Value

		group.Put( base.Assetses().DBKey(), tx.From.Bytes(), astfrom.Encode() )
		group.Put( base.Assetses().DBKey(), tx.To.Bytes(), astto.Encode() )

		group.Put(
			base.Receipts().DBKey(),
			receiptKey,
			ARsp.RawSusccessResponse(block.TxConfirm),
			)

		return nil
	}

	return nil
}