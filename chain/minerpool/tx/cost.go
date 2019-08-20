package workflow

import (
	"errors"
	ACComm "github.com/ayachain/go-aya/chain/common"
	"github.com/ayachain/go-aya/vdb"
	AAsset "github.com/ayachain/go-aya/vdb/assets"
	ARsp "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
)

func DoCostHandle( tx *ATx.Transaction, base vdb.CacheCVFS, txindex int ) error {

	if tx.Verify() {

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
		if astfrom.Avail < ATx.StaticCostMapping[tx.Type] || astfrom.Vote < ATx.StaticCostMapping[tx.Type] {

			base.Receipts().Put( txHash, tx.BlockIndex, ARsp.ExpectedReceipt(ACComm.TxInsufficientFunds, nil).Encode() )

			return errors.New( "insufficient funds" )
		}

		// success
		astfrom.Avail -= ATx.StaticCostMapping[tx.Type]
		astfrom.Vote -= ATx.StaticCostMapping[tx.Type]

		// cost
		superNodes := base.Nodes().GetSuperNodeList()
		recvAddr := superNodes[txindex % len(superNodes)].Owner
		costRecver, _ := base.Assetses().AssetsOf(recvAddr)
		if costRecver == nil {

			costRecver = &AAsset.Assets{Version: AAsset.DRVer, Avail: ATx.StaticCostMapping[tx.Type], Vote: ATx.StaticCostMapping[tx.Type], Locked: 0}

		} else {

			costRecver.Avail += ATx.StaticCostMapping[tx.Type]
			costRecver.Vote += ATx.StaticCostMapping[tx.Type]

		}

		base.Assetses().Put(recvAddr, costRecver)
		base.Assetses().Put(tx.From, astfrom)

		return nil

	}

	return errors.New("tx sig verify expected")
}