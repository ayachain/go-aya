package block

import (
	"context"
	"encoding/json"
	AMsgBlock "github.com/ayachain/go-aya/chain/message/block"
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ayachain/go-aya/consensus/impls/APOS/in/block/workflow"
	Avdb "github.com/ayachain/go-aya/vdb"
	AvdbTx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/pkg/errors"
)


var (
	ErrExpected 			= errors.New("lock or unlock amount expected")
	ErrTxVerifyExpected 	= errors.New("transaction verify failed")
	ErrParmasExpected		= errors.New("transaction params expected")
	ErrMessageTypeExped  	= errors.New("message type expected")

	TxConfirm              	= []byte("Confirm")
	TxReceiptsNotenough 	= []byte("not enough avail or voting balance")
)

func WokerSwitcher( msg interface{}, vdb Avdb.CVFS, ind *core.IpfsNode, group *AWork.TaskBatchGroup ) (interface{}, error) {

	rawblock, ok := msg.(*AMsgBlock.MsgRawBlock)
	if !ok {
		return nil, ErrMessageTypeExped
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
	if err := json.Unmarshal(iblock.RawData(), txlist); err != nil {
		return nil, err
	}

	for _, tx := range txlist {

		switch tx.Data {
		case []byte("UNLOCK"), []byte("LOCK") :
			if err := workflow.DoLockAmount(tx, group, vdb); err != nil {
				return nil, err
			}

		default:
			if err := workflow.DoTransfer(tx, group, vdb); err != nil {
				return nil, err
			}
		}

	}

	return nil, nil
}