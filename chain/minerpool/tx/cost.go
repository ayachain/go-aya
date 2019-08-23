package workflow

import (
	"errors"
	ACComm "github.com/ayachain/go-aya/chain/common"
	"github.com/ayachain/go-aya/vdb"
	"github.com/ayachain/go-aya/vdb/im"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/ethereum/go-ethereum/common"
)

func DoCostHandle( tx *im.Transaction, base vdb.CacheCVFS, txindex int ) error {

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
		if astfrom.Avail < ATx.StaticCostMapping[tx.Type] || astfrom.Vote < ATx.StaticCostMapping[tx.Type] {

			base.Receipts().Put( txHash, tx.BlockIndex, im.ExpectedReceipt(ACComm.TxInsufficientFunds, nil) )

			return errors.New( "insufficient funds" )
		}

		// success
		astfrom.Avail -= ATx.StaticCostMapping[tx.Type]
		astfrom.Vote -= ATx.StaticCostMapping[tx.Type]

		// cost
		superNodes := base.Nodes().GetSuperNodeList()
		recvAddr := superNodes[txindex % len(superNodes)].Owner
		costRecver, _ := base.Assetses().AssetsOf( common.BytesToAddress(recvAddr) )
		if costRecver == nil {

			costRecver = &im.Assets{ Avail: ATx.StaticCostMapping[tx.Type], Vote: ATx.StaticCostMapping[tx.Type], Locked: 0 }

		} else {

			costRecver.Avail += ATx.StaticCostMapping[tx.Type]
			costRecver.Vote += ATx.StaticCostMapping[tx.Type]

		}

		base.Assetses().Put( common.BytesToAddress(recvAddr), costRecver)
		base.Assetses().Put( common.BytesToAddress(tx.From), astfrom)

		return nil

	}

	return errors.New("tx sig verify expected")
}