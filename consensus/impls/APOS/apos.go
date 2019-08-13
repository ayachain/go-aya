package APOS

import (
	"context"
	ACore "github.com/ayachain/go-aya/consensus/core"
	AGroup "github.com/ayachain/go-aya/consensus/core/worker"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/common"
	"github.com/ayachain/go-aya/consensus/impls/APOS/history"
	"github.com/ayachain/go-aya/consensus/impls/APOS/workflow"
	"github.com/ayachain/go-aya/vdb"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	ACInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	AMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMined "github.com/ayachain/go-aya/vdb/minined"
	ARsp "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-ipfs/core"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const DeveloperMode = false

type APOSConsensusNotary struct {

	ACore.Notary

	ind *core.IpfsNode

	hst *history.History

	mu sync.Mutex
}

func NewAPOSConsensusNotary( ind *core.IpfsNode ) *APOSConsensusNotary {

	notary := &APOSConsensusNotary{
		ind:ind,
		hst:history.New(),
	}

	return notary
}


func (n *APOSConsensusNotary) MiningBlock( block *AMBlock.MBlock, cvfs vdb.CacheCVFS, txs []*ATx.Transaction ) (*AGroup.TaskBatchGroup, error) {

	if txs == nil {

		ctx, cancel := context.WithTimeout( context.TODO(), time.Second * 32 )

		txs = block.ReadTxsFromDAG(ctx, n.ind)

		cancel()

		if txs == nil || len(txs) <= 0 {
			return nil, errors.New("read mining block timeout")
		}

	}

	for i, tx := range txs {

		// is transaction override
		txc, err := cvfs.Transactions().GetTxCount(tx.From)

		if err != nil {
			continue
		}

		if tx.Tid < txc {
			cvfs.Receipts().Put(tx.GetHash256(), block.Index, ARsp.ExpectedReceipt(APosComm.TxOverrided, nil).Encode())
			continue
		}

		// Handle Cost
		if workflow.DoCostHandle( tx, cvfs, i ) != nil {
			continue
		}

		cvfs.Transactions().Put(tx, block.Index)

		switch tx.Type {

		//case "UNLOCK", "LOCK":
		//	if err := workflow.DoLockAmount(tx, group, vdb); err != nil {
		//		return nil, err
		//	}

		case ATx.NormalTransfer :

			if err := workflow.DoTransfer( tx, cvfs ); err != nil {
				return nil, err
			}

		default:

			cvfs.Receipts().Put(tx.GetHash256(), block.Index, ARsp.ExpectedReceipt(APosComm.TxUnsupportTransactionType, nil).Encode())

		}
	}

	// should pos block
	//workflow.CanPos( block, cvfs )

	return cvfs.MergeGroup(), nil
}

func (n *APOSConsensusNotary) NewBlockHasConfirm(  ) {

	n.mu.Lock()
	defer n.mu.Unlock()

	n.hst.Clear()

}

func (n *APOSConsensusNotary) TrustOrNot( msg *pubsub.Message, mtype ACore.NotaryMessageType, cvfs vdb.CVFS ) <- chan bool {

	replayChan := make(chan bool)

	if DeveloperMode {

		go func() {
			replayChan <- true
		}()

		return replayChan
	}

	go func() {

		n.mu.Lock()
		defer n.mu.Unlock()

		sender, err := cvfs.Nodes().GetNodeByPeerId(msg.GetFrom().String())

		if err != nil {
			replayChan <- false
		}

		if sender == nil {
			replayChan <- false
			return
		}

		msgHash := crypto.Keccak256Hash(msg.Data)

		threshold := (cvfs.Nodes().GetSuperMaterTotalVotes() * 51) / 100

		switch mtype {

		case ACore.NotaryMessageChainInfo:

			if msg.Data[0] == ACInfo.MessagePrefix {

				replayChan <- n.hst.CanConsensus(msgHash.String(), sender, threshold)

			}

		case ACore.NotaryMessageTransaction:

			if msg.Data[0] == ATx.MessagePrefix {

				//log.Infof("ListenMessage : %v -> %v (Transaction)", sender.PeerID, msgHash.String())

				tx := &ATx.Transaction{}

				if err := tx.RawMessageDecode(msg.Data); err != nil {
					replayChan <- false
				}

				replayChan <- tx.Verify()
			}

		case ACore.NotaryMessageMiningBlock:

			if msg.Data[0] == AMBlock.MessagePrefix {

				//log.Infof("ListenMessage : %v -> %v (MiningBlock)", sender.PeerID, msgHash.String())

				replayChan <- n.hst.CanConsensus(msgHash.String(), sender, threshold)

			}

		case ACore.NotaryMessageConfirmBlock:

			if msg.Data[0] == ABlock.MessagePrefix {

				//log.Infof("ListenMessage : %v -> %v (ConfirmBlock)", sender.PeerID, msgHash.String())

				replayChan <- n.hst.CanConsensus(msgHash.String(), sender, threshold)

			}

		case ACore.NotaryMessageMinedRet:

			if msg.Data[0] == AMined.MessagePrefix {

				//log.Infof("ListenMessage : %v -> %v (MinedReceipt)", sender.PeerID, msgHash.String())

				replayChan <- n.hst.CanConsensus(msgHash.String(), sender, threshold)

			}

		default:
			replayChan <- false
		}

		return

	}()

	return replayChan
}