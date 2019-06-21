package APOS

import (
	"context"
	"encoding/json"
	ACore "github.com/ayachain/go-aya/consensus/core"
	AGroup "github.com/ayachain/go-aya/consensus/core/worker"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/common"
	"github.com/ayachain/go-aya/consensus/impls/APOS/workflow"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ayachain/go-aya/vdb"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	ARsp "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/syndtr/goleveldb/leveldb"
)

type APOSConsensusNotary struct {

	ACore.Notary

	workctx    context.Context

	workCancel context.CancelFunc

	ind *core.IpfsNode

	db *leveldb.DB
}

func NewAPOSConsensusNotary( ind *core.IpfsNode ) *APOSConsensusNotary {

	ctx, cancel := context.WithCancel(context.Background())

	notary := &APOSConsensusNotary{
		workctx:ctx,
		workCancel:cancel,
		ind:ind,
	}

	return notary
}

func (n *APOSConsensusNotary) MiningBlock( block *AMsgMBlock.MBlock, cvfs vdb.CacheCVFS ) (*AGroup.TaskBatchGroup, error) {

	txsCid, err := cid.Decode(block.Txs)
	if err != nil {
		return nil, err
	}

	iblock, err := n.ind.Blocks.GetBlock(context.TODO(), txsCid)
	if err != nil {
		return nil, err
	}

	txlist := make([]*ATx.Transaction, block.Txc)
	if err := json.Unmarshal(iblock.RawData(), &txlist); err != nil {
		return nil, err
	}

	for _, tx := range txlist {


		// is transaction override
		txc, err := cvfs.Transactions().GetTxCount(tx.From)
		if err != nil {
			continue
		}

		if tx.Tid < txc {
			cvfs.Receipts().Put(tx.GetHash256(), block.Index, ARsp.ExpectedReceipt(APosComm.TxOverrided, nil).Encode())
			continue
		}

		cvfs.Transactions().Put(tx, block.Index)

		switch string(tx.Data) {

		//case "UNLOCK", "LOCK":
		//	if err := workflow.DoLockAmount(tx, group, vdb); err != nil {
		//		return nil, err
		//	}

		default:

			if err := workflow.DoTransfer(tx, cvfs); err != nil {
				return nil, err
			}

		}

	}

	return cvfs.MergeGroup(), nil
}

func (n *APOSConsensusNotary) TrustOrNot( msg *AKeyStore.ASignedRawMsg, mtype ACore.NotaryMessageType ) <- chan bool {

	replayChan := make(chan bool)

	go func() {

		//from, err := msg.ECRecover()
		//if err != nil {
		//	replayChan <- false
		//	return
		//}

		switch mtype {

		case ACore.NotaryMessageChainInfo:

		case ACore.NotaryMessageTransaction:

		case ACore.NotaryMessageMiningBlock:

		case ACore.NotaryMessageConfirmBlock:

		case ACore.NotaryMessageReceipt:

		}

	}()

	return replayChan
}