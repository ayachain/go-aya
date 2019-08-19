package chain

import (
	"context"
	"errors"
	AMinerPool "github.com/ayachain/go-aya/chain/minerpool"
	ASDaemon "github.com/ayachain/go-aya/chain/sdaemon/common"
	"github.com/ayachain/go-aya/chain/txpool"
	"github.com/ayachain/go-aya/consensus/core/worker"
	AMsgCenter "github.com/ayachain/go-aya/consensus/msgcenter"
	"github.com/ayachain/go-aya/vdb"
	ACBlock "github.com/ayachain/go-aya/vdb/block"
	ACInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	AIndexs "github.com/ayachain/go-aya/vdb/indexes"
	AMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMinied "github.com/ayachain/go-aya/vdb/minined"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/pin"
	"github.com/prometheus/common/log"
	"time"
)

var(
	ErrAlreadyExistConnected		= errors.New("chan already exist connected")
	ErrCantLinkToChainExpected		= errors.New("not found chain in Aya")
	ErrMergeFailed					= errors.New("CVFS merge failed")
	ErrMergeInvalidChainID			= errors.New("invalid chain id")
	ErrMergeInvalidLatest			= errors.New("invalid latest block")
	ErrMergeNonlinear				= errors.New("nonlinear merge batch ")
	ErrMergeUnlinkCVFS				= errors.New("can not link target CVFS")
)

const  AChainMapKey = "aya.chain.list.map"

type Chain interface {

	Disconnect()

	CVFServices() vdb.CVFS

	GetTxPool() txpool.TxPool
}

type aChain struct {

	Chain

	ChainId string

	INode *core.IpfsNode

	CVFS vdb.CVFS

	IDX AIndexs.IndexesServices

	TXP txpool.TxPool

	AMC AMsgCenter.MessageCenter

	AMP AMinerPool.MinerPool

	ASD ASDaemon.StatDaemon

	CancelCh chan struct{}
}

func (chain *aChain) LinkStart( ctx context.Context ) error {

	go func() {

		sctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for {

			select {
			case tmsg := <- chain.AMC.TrustMessage():
				chain.TrustMessageSwitcher(sctx, tmsg)

			case <- ctx.Done():
				return
			}

		}

	}()

	txPoolCtx, txpoolCancel := context.WithCancel(ctx)
	defer txpoolCancel()

	amcCtx, amcCancel := context.WithCancel(ctx)
	defer amcCancel()

	asdCtx, asdCancel := context.WithCancel(ctx)
	defer asdCancel()

	go chain.AMC.PowerOn(amcCtx, chain.ChainId, chain.INode)
	go chain.ASD.PowerOn(asdCtx)
	go chain.TXP.PowerOn(txPoolCtx)

	select {
	case <- chain.CancelCh:
		return nil

	case <- txPoolCtx.Done():
		return nil

	case <- amcCtx.Done():
		return nil

	case <- ctx.Done():
		return nil
	}
}

func (chain *aChain) Disconnect() {
	chain.CancelCh <- struct{}{}
}

func (chain *aChain) GetTxPool() txpool.TxPool {
	return chain.TXP
}

func (chain *aChain) CVFServices() vdb.CVFS {
	return chain.CVFS
}

func (chain *aChain) TrustMessageSwitcher( ctx context.Context, msg []byte ) {

	switch msg[0] {

	case AMBlock.MessagePrefix:

		mblock := &AMBlock.MBlock{}
		if err := mblock.Decode(msg[1:]); err != nil {
			return
		}

		chain.ASD.SendingSignal( mblock.Index, ASDaemon.SignalDoMining )

		go func() {

			mret := chain.AMP.PutTask(ctx, AMinerPool.NewTask( mblock ), time.Second * 60)
			if mret.Err != nil {
				log.Warn(mret.Err)
				return
			}

			if err := chain.AMC.PublishMessage( &AMinied.Minined {
				MBlock:mblock,
				Batcher:mret.Batcher.Upload(chain.INode),
			}, AMsgCenter.GetChannelTopics(mblock.ChainID, AMsgCenter.MessageChannelBatcher) ); err != nil {
				log.Warn(err)
			}

			chain.ASD.SendingSignal( mblock.Index, ASDaemon.SignalDoReceipting )

			return

		}()


	case AMinied.MessagePrefix:

		batcher := &AMinied.Minined{}
		if err := batcher.Decode(msg[1:]); err != nil {
			return
		}

		if cinfo, err := chain.ForkMergeBatch(ctx, batcher); err != nil {
			log.Warn(err)
			return

		} else {

			if err := chain.AMC.PublishMessage( cinfo, AMsgCenter.GetChannelTopics(batcher.MBlock.ChainID, AMsgCenter.MessageChannelAppend)); err != nil {
				log.Warn(err)
				return
			}

			chain.ASD.SendingSignal( batcher.MBlock.Index, ASDaemon.SignalDoConfirming )
		}

		return

	case ACInfo.MessagePrefix:

		cinfo := &ACInfo.ChainInfo{}
		if err := cinfo.Decode(msg[1:]); err != nil {
			return
		}

		/// check chain id
		if cinfo.ChainID != chain.ChainId {
			return
		}

		/// check block index
		lidx, err := chain.IDX.GetLatest()
		if err != nil || lidx == nil {

			/// if this chain's current index services has error, use new chain info.
			if err := chain.IDX.SyncToCID(cinfo.FinalCVFS); err != nil {
				panic(err)
			}

			chain.ASD.SendingSignal( cinfo.BlockIndex, ASDaemon.SignalOnConfirmed )
			return
		}

		/// local node's chain is longer
		if lidx.BlockIndex >= cinfo.BlockIndex {
			return

		} else {

			/// chain info's chain is longer use it
			if err := chain.IDX.SyncToCID(cinfo.Indexes); err != nil {
				log.Error(err)
				return
			}

		}

		/// appended a new block success, pin this block create's new data
		lidx, err = chain.IDX.GetLatest()
		if err != nil {
			panic(err)
		}

		lblock, err := chain.CVFS.Blocks().GetLatestBlock()
		if err != nil {
			panic(err)
		}

		chain.INode.Pinning.PinWithMode( lblock.Txs, pin.Any )
	}

}

func (chain *aChain) ForkMergeBatch( ctx context.Context, mret *AMinied.Minined ) (*ACInfo.ChainInfo, error) {

	if mret.MBlock.ChainID != chain.ChainId {
		return nil, ErrMergeInvalidChainID
	}

	lidx, err := chain.IDX.GetLatest()
	if err != nil {
		return nil, ErrMergeInvalidLatest
	}

	if mret.MBlock.Index - lidx.BlockIndex != 1 {
		return nil, ErrMergeNonlinear
	}

	/// Read batcher
	rctx, cancel := context.WithTimeout(ctx, time.Second * 32)
	batcher := taskBatchGroupFromCID(rctx, chain.INode, mret.Batcher)
	defer cancel()
	if rctx.Err() != nil {
		return nil, rctx.Err()
	}

	/// Create confirm block
	cblock := mret.MBlock.Confirm( mret.Batcher.String() )
	if cblock == nil {
		return nil, ErrMergeFailed
	}

	/// Append confirm block
	batcher.Put(ACBlock.DBPath, cblock.GetHash().Bytes(), cblock.Encode() )

	/// try merge
	ccid, err := chain.CVFS.ForkMergeBatch(batcher)
	if err != nil {
		return nil, ErrMergeFailed
	}

	/// try fork merge
	idxfcid, err := AIndexs.ForkMerge(chain.INode, chain.ChainId, &AIndexs.Index{
		BlockIndex:mret.MBlock.Block.Index,
		Hash:cblock.GetHash(),
		FullCID:ccid,
	})

	if err != nil || idxfcid == cid.Undef {
		return nil, err
	}

	/// create chain info waiting review
	finfo := &ACInfo.ChainInfo{
		ChainID:mret.MBlock.ChainID,
		Indexes:idxfcid,
	}

	return finfo, nil
}

func taskBatchGroupFromCID ( ctx context.Context, ind *core.IpfsNode, c cid.Cid) *worker.TaskBatchGroup {

	reply := make(chan *worker.TaskBatchGroup)
	defer close(reply)

	go func() {

		blk, err := ind.Blocks.GetBlock(ctx, c)
		if err != nil {
			reply <- nil
			return
		}

		if blk == nil || blk.RawData() == nil {
			reply <- nil
			return
		}

		batch := &worker.TaskBatchGroup{}
		if err := batch.Decode(blk.RawData()); err != nil {
			reply <- nil
			return
		} else {

			reply <- batch
			return
		}

	}()

	select {
	case <- ctx.Done():
		return nil

	case b := <- reply:
		return b
	}
}
