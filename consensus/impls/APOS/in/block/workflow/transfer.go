package workflow

import (
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/in/common"
	ARsp "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb"
	AAsset "github.com/ayachain/go-aya/vdb/assets"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/pkg/errors"
)

var expectedErr = errors.New("transfer expected")

func DoTransfer( tx *ATx.Transaction, base vdb.CacheCVFS ) error {

	if !tx.Verify() {

		txHash := tx.GetHash256()

		astfrom, err := base.Assetses().AssetsOf(tx.From)
		if err != nil {
			return nil
		}

		astto, err := base.Assetses().AssetsOf(tx.To)
		if err != nil || astto == nil {
			astto = &AAsset.Assets{ Version:AAsset.DRVer,Avail:0,Vote:0,Locked:0 }
		}

		// expected
		if astfrom.Avail < tx.Value || astfrom.Vote < tx.Value {

			base.Receipts().Put( txHash, tx.BlockIndex, ARsp.RawExpectedResponse(APosComm.TxReceiptsNotenough) )

			return nil
		}

		// success
		astfrom.Avail -= tx.Value
		astfrom.Vote -= tx.Value

		astto.Avail += tx.Value
		astto.Vote += tx.Value

		base.Assetses().Put( tx.From, astfrom )
		base.Assetses().Put( tx.To, astto )
		base.Receipts().Put( txHash, tx.BlockIndex, ARsp.RawExpectedResponse(APosComm.TxConfirm) )

		return nil
	}

	return nil
}