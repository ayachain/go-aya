package workflow

import (
	ACComm "github.com/ayachain/go-aya/chain/common"
	"github.com/ayachain/go-aya/vdb"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ethereum/go-ethereum/common"
)

func DoTransfer( tx *im.Transaction, base vdb.CacheCVFS ) error {

	if tx.Verify() {

		txHash := tx.GetHash256()

		astfrom, err := base.Assetses().AssetsOf( common.BytesToAddress(tx.From) )
		if err != nil {
			return nil
		}

		astto, err := base.Assetses().AssetsOf( common.BytesToAddress(tx.To) )
		if err != nil || astto == nil {
			astto = &im.Assets{ Avail:0,Vote:0,Locked:0 }
		}

		// expected
		if astfrom.Avail < tx.Value || astfrom.Vote < tx.Value {

			base.Receipts().Put( txHash, tx.BlockIndex, im.ExpectedReceipt(ACComm.TxInsufficientFunds, nil) )

			return nil
		}

		// success
		astfrom.Avail -= tx.Value
		astfrom.Vote -= tx.Value

		astto.Avail += tx.Value
		astto.Vote += tx.Value

		base.Assetses().Put( common.BytesToAddress(tx.From), astfrom )
		base.Assetses().Put( common.BytesToAddress(tx.To), astto )
		base.Receipts().Put( txHash, tx.BlockIndex, im.ConfirmReceipt(ACComm.TxConfirm, nil) )

		return nil
	}

	return nil
}