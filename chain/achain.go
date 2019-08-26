package chain

import (
	"context"
	"errors"
	AMinerPool "github.com/ayachain/go-aya/chain/minerpool"
	AMsgCenter "github.com/ayachain/go-aya/chain/msgcenter"
	ASDaemon "github.com/ayachain/go-aya/chain/sdaemon/common"
	"github.com/ayachain/go-aya/chain/txpool"
	"github.com/ayachain/go-aya/vdb"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	ACBlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ayachain/go-aya/vdb/im"
	AIndexs "github.com/ayachain/go-aya/vdb/indexes"
	VDBMerge "github.com/ayachain/go-aya/vdb/merger"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
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
)

const AChainMapKey = "AChainMapKey"

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

func (chain *aChain) TrustMessageSwitcher( ctx context.Context, msg interface{} ) {

	switch msg.(type) {

	case *im.Block:

		mblock := msg.(*im.Block)

		chain.ASD.SendingSignal( mblock.Index, ASDaemon.SignalDoMining )

		go func() {

			st := time.Now().UnixNano()
			log.Info("Starting Mining")
			defer func() {
				log.Info("End Mining : %v ms", float64(time.Now().UnixNano() - st) / float64(1e6) )
			}()

			mret := chain.AMP.PutTask( AMinerPool.NewTask( mblock ) )
			if mret.Err != nil {
				log.Warn(mret.Err)
				chain.ASD.SendingSignal( mblock.Index, ASDaemon.SignalInterrupt )
				return
			}

			chain.ASD.SendingSignal( mblock.Index, ASDaemon.SignalDoReceipting )
			if err := chain.AMC.PublishMessage( &im.Minined {
				MBlock:mblock,
				Batcher:mret.Batcher.Upload(chain.INode).Bytes(),
			}, AMsgCenter.GetChannelTopics(mblock.ChainID, AMsgCenter.MessageChannelMined) ); err != nil {
				log.Warn(err)
			}

			return

		}()


	case *im.Minined:

		mined := msg.(*im.Minined)

		st := time.Now().UnixNano()
		log.Info("Starting ForkMerge")
		defer func() {
			log.Info("End ForkMerge : %v ms", float64(time.Now().UnixNano() - st) / float64(1e6) )
		}()

		sctx, cancel := context.WithTimeout(ctx, time.Second * 32)
		defer cancel()

		if cinfo, err := chain.ForkMergeBatch(sctx, mined); err != nil {

			log.Warn(err)
			return

		} else {

			if err := chain.AMC.PublishMessage( cinfo, AMsgCenter.GetChannelTopics(mined.MBlock.ChainID, AMsgCenter.MessageChannelChainInfo)); err != nil {
				log.Warn(err)
				return
			}

			chain.ASD.SendingSignal( mined.MBlock.Index, ASDaemon.SignalDoConfirming )
		}

		return

	case *im.ChainInfo:

		cinfo := msg.(*im.ChainInfo)

		st := time.Now().UnixNano()
		log.Info("Starting ConfirmChainInfo")
		defer func() {
			log.Info("End ConfirmChainInfo : %v ms", float64(time.Now().UnixNano() - st) / float64(1e6) )
		}()

		/// check chain id
		if cinfo.ChainID != chain.ChainId {
			return
		}

		/// check block index
		lidx, err := chain.IDX.GetLatest()
		if err != nil || lidx == nil {

			/// if this chain's current index services has error, use new chain info.
			ccid, err := cid.Cast(cinfo.FinalCVFS)
			if err != nil {
				log.Error(err)
				return
			}

			if err := chain.IDX.SyncToCID( ccid ); err != nil {
				panic(err)
			}

			chain.ASD.SendingSignal( cinfo.BlockIndex, ASDaemon.SignalOnConfirmed )
			return
		}

		/// local node's chain is longer
		if lidx.BlockIndex >= cinfo.BlockIndex {
			return

		} else {

			ccid, err := cid.Cast(cinfo.Indexes)
			if err != nil {
				log.Error(err)
				return
			}

			/// chain info's chain is longer use it
			if err := chain.IDX.SyncToCID(ccid); err != nil {
				log.Error(err)
				return
			}

			chain.ASD.SendingSignal( cinfo.BlockIndex, ASDaemon.SignalOnConfirmed )
			return
		}
	}

}

func (chain *aChain) ForkMergeBatch( ctx context.Context, mret *im.Minined ) (*im.ChainInfo, error) {

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
	ccid, err := cid.Cast(mret.Batcher)
	if err != nil {
		return nil, err
	}

	rctx, cancel := context.WithTimeout(ctx, time.Second * 32)
	merger := ReadMergeFromCID(rctx, chain.INode, ccid)
	defer cancel()
	if rctx.Err() != nil {
		return nil, rctx.Err()
	}

	/// Create confirm block
	ccid, err = cid.Cast(mret.Batcher)
	if err != nil {
		return nil, err
	}

	cblock := ABlock.ConfirmBlock(mret.MBlock, ccid )
	if cblock == nil {
		return nil, ErrMergeFailed
	}

	/// Append confirm block
	bbc, err := proto.Marshal(cblock)
	if err != nil {
		return nil, err
	}
	merger.Put(ACBlock.DBPath, cblock.GetHash().Bytes(), bbc )

	/// try merge
	ccid, err = chain.CVFS.ForkMergeBatch(merger)
	if err != nil {
		return nil, ErrMergeFailed
	}

	/// try fork merge
	idxfcid, err := AIndexs.ForkMerge(chain.INode, chain.ChainId, &AIndexs.Index{
		BlockIndex:mret.MBlock.Index,
		Hash:cblock.GetHash(),
		FullCID:ccid,
	})

	if err != nil || idxfcid == cid.Undef {
		return nil, err
	}

	/// create chain info waiting review
	finfo := &im.ChainInfo{
		ChainID:mret.MBlock.ChainID,
		Indexes:idxfcid.Bytes(),
		BlockIndex:mret.MBlock.Index,
		FinalCVFS:ccid.Bytes(),
	}

	return finfo, nil
}

func ReadMergeFromCID ( ctx context.Context, ind *core.IpfsNode, c cid.Cid) VDBMerge.CVFSMerger {

	reply := make(chan VDBMerge.CVFSMerger)
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

		merger := VDBMerge.NewMerger()
		if err := merger.Load(blk.RawData()); err != nil {

			reply <- nil
			return

		} else {

			reply <- merger
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
