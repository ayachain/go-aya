package worker

import (
	"encoding/binary"
	AWorker "github.com/ayachain/go-aya/consensus/core/worker"
	ARsp "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/pkg/errors"
)

var expectedErr = errors.New("SimpleWorking expected")

var (
	TxConfirm              = []byte("Confirm")
	TxerrreceiptsNotenough = []byte("not enough avail or voting balance")
)

func transfer( tx *ATx.Transaction, group *AWorker.TaskBatchGroup, base vdb.CVFS ) error {

	if !tx.Verify() {

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
				ARsp.RawExpectedResponse(TxerrreceiptsNotenough),
				)

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
			ARsp.RawSusccessResponse(TxConfirm),
			)

	}

	return nil
}