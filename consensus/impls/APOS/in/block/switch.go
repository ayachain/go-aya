package block

import (
	"context"
	"encoding/json"
	"github.com/ayachain/go-aya/consensus/impls/APOS/in/block/workflow"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/in/common"
	Avdb "github.com/ayachain/go-aya/vdb"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AvdbTx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
)

func WokerSwitcher( msg interface{}, vdb Avdb.CacheCVFS, ind *core.IpfsNode ) (interface{}, error) {

	rawblock, ok := msg.(*AMsgMBlock.MBlock)
	if !ok {
		return nil, APosComm.ErrMessageTypeExped
	}

	txsCid, err := cid.Decode(rawblock.Txs)
	if err != nil {
		return nil, err
	}

	iblock, err := ind.Blocks.GetBlock(context.TODO(), txsCid)
	if err != nil {
		return nil, err
	}

	txlist := make([]*AvdbTx.Transaction, rawblock.Txc)
	if err := json.Unmarshal(iblock.RawData(), &txlist); err != nil {
		return nil, err
	}

	for _, tx := range txlist {

		txc, err := vdb.Transactions().GetTxCount(tx.From)
		if err != nil {
			continue
		}

		if txc != tx.Tid - 1 {
			continue
		}

		switch string(tx.Data) {

		//case "UNLOCK", "LOCK":
		//	if err := workflow.DoLockAmount(tx, group, vdb); err != nil {
		//		return nil, err
		//	}

		default:

			if err := workflow.DoTransfer(tx, vdb); err != nil {
				return nil, err
			}

		}

		vdb.Transactions().Put(tx, rawblock.Index)
	}

	return msg, nil
}