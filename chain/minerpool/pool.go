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

	PutTask( ctx context.Context, task *MiningTask, timeLimit time.Duration ) *MiningResult

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

func (mp *aMinerPool) PutTask( ctx context.Context, task *MiningTask, timeLimit time.Duration ) *MiningResult {

	mctx, mcancel := context.WithTimeout(ctx, timeLimit)
	defer mcancel()

	reply := make(chan *MiningResult)

	go func( ctx context.Context, task *MiningTask ) {

		/// Compare chainid
		if task.MiningBlock.ChainID != mp.chainID {
			reply <- &MiningResult{
				Err:ErrInvalidChainID,
				Task:task,
				Batcher:nil,
			}
			return
		}

		/// Read latest block index
		lidx, err := mp.idxs.GetLatest()
		if err != nil {
			reply <- &MiningResult{
				Err:ErrInvalidLatest,
				Task:task,
				Batcher:nil,
			}
			return
		}

		/// Compare block index and parent hash
		if task.MiningBlock.Index - lidx.BlockIndex != 1 || task.MiningBlock.Parent != lidx.Hash.String() {
			reply <- &MiningResult{
				Err:ErrNotLinearBlock,
				Task:task,
				Batcher:nil,
			}
			return
		}

		/// Read tx list from IPFS dag services
		txsCtx, txsCancel := context.WithTimeout(ctx, 16)
		defer txsCancel()

		task.Txs = task.MiningBlock.ReadTxsFromDAG(txsCtx, mp.ind)
		if txsCtx.Err() != nil {
			reply <- &MiningResult{
				Err:ErrReadTxsTimeOut,
				Task:task,
				Batcher:nil,
			}
			return
		}

		/// Create chain data vdb services
		cvfs, err := vdb.LinkVFS(task.MiningBlock.ChainID, mp.ind, mp.idxs)
		if err != nil {

			reply <- &MiningResult{
				Err:ErrLinkCVFS,
				Task:task,
				Batcher:nil,
			}

			return
		}

		/// Create cvfs writer
		vwriter, err := cvfs.NewCVFSWriter()
		if err != nil {

			reply <- &MiningResult{
				Err:ErrCreateCVFSCache,
				Task:task,
				Batcher:nil,
			}

			return
		}
		defer vwriter.Close()

		/// payload mining task object
		task.VWriter = vwriter

		miner := newMiner()

		reply <- miner.DoTask(ctx, task)

		return

	}( mctx, task )

	var err error

	select {
	case <- ctx.Done():
		err = ctx.Err()
		goto ErrorReply

	case <- mctx.Done():
		err = ctx.Err()
		goto ErrorReply

	case ret := <- reply:
		return ret
	}

ErrorReply:

	return &MiningResult{
		Err:err,
		Task:task,
		Batcher:nil,
	}
}
