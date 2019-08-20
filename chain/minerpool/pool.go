package minerpool

import (
	"context"
	ASD "github.com/ayachain/go-aya/chain/sdaemon/common"
	"github.com/ayachain/go-aya/vdb"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
	"github.com/ipfs/go-ipfs/core"
	"time"
)

type MinerPool interface {
	PutTask( task *MiningTask ) *MiningResult
}


type aMinerPool struct {

	MinerPool

	ind *core.IpfsNode

	idxs AIndexes.IndexesServices

	chainID string

	asd ASD.StatDaemon
}

func NewPool( chainID string, ind *core.IpfsNode, idxser AIndexes.IndexesServices, asd ASD.StatDaemon ) MinerPool {

	return &aMinerPool{
		ind:ind,
		idxs:idxser,
		chainID:chainID,
		asd:asd,
	}

}

func (mp *aMinerPool) PutTask( task *MiningTask ) *MiningResult {

	var err error

	/// Compare chainid
	if task.MiningBlock.ChainID != mp.chainID {
		return &MiningResult{
			Err:ErrInvalidChainID,
			Task:task,
			Batcher:nil,
		}
	}

	/// Read latest block index
	lidx, err := mp.idxs.GetLatest()
	if err != nil {
		return &MiningResult{
			Err:ErrInvalidLatest,
			Task:task,
			Batcher:nil,
		}
	}

	/// Compare block index and parent hash
	if task.MiningBlock.Index - lidx.BlockIndex != 1 || task.MiningBlock.Parent != lidx.Hash.String() {
		return &MiningResult{
			Err:ErrNotLinearBlock,
			Task:task,
			Batcher:nil,
		}
	}

	/// Read tx list from IPFS dag services
	txsCtx, txsCancel := context.WithTimeout(context.TODO(), time.Second * 16)
	defer txsCancel()

	task.Txs = task.MiningBlock.ReadTxsFromDAG(txsCtx, mp.ind)
	if txsCtx.Err() != nil {
		return &MiningResult{
			Err:ErrReadTxsTimeOut,
			Task:task,
			Batcher:nil,
		}
	}

	/// Create chain data vdb services
	cvfs, err := vdb.LinkVFS(task.MiningBlock.ChainID, mp.ind, mp.idxs)
	if err != nil {
		return &MiningResult{
			Err:ErrLinkCVFS,
			Task:task,
			Batcher:nil,
		}
	}

	/// Create cvfs writer
	vwriter, err := cvfs.NewCVFSWriter()
	if err != nil {
		return &MiningResult{
			Err:ErrCreateCVFSCache,
			Task:task,
			Batcher:nil,
		}
	}
	defer vwriter.Close()

	/// payload mining task object
	task.VWriter = vwriter

	return mp.doTask(task)
}
