package workflow

import (
	"encoding/binary"
	AWorker "github.com/ayachain/go-aya/consensus/core/worker"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/in/common"
	ARsp "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb"
	AAsset "github.com/ayachain/go-aya/vdb/assets"
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

			group.Put(
				base.Receipts().DBKey(),
				receiptKey,
				ARsp.RawExpectedResponse(APosComm.TxReceiptsNotenough),
			)

			return nil
		}

		astto, err := base.Assetses().AssetsOf(tx.To.Bytes())
		if err != nil || astto == nil {
			astto = &AAsset.Assets{ Version:AAsset.DRVer,Avail:0,Vote:0,Locked:0 }
		}

		// expected
		if astfrom.Avail < tx.Value || astfrom.Vote < tx.Value {

			group.Put(
				base.Receipts().DBKey(),
				receiptKey,
				ARsp.RawExpectedResponse(APosComm.TxReceiptsNotenough),
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
			ARsp.RawSusccessResponse(APosComm.TxConfirm),
			)

		return nil
	}

	return nil
}