package txpool

import (
	"context"
	"fmt"
	"github.com/ayachain/go-aya/consensus/core"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AChainInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	"strings"
	"time"
)

func syncListener(ctx context.Context ) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadSyncer)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans.Store(ATxPoolThreadSyncer, make(chan []byte, ATxPoolThreadSyncerBuff))

	subCtx, subCancel := context.WithCancel(ctx)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans.Load( ATxPoolThreadSyncer )
		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete( ATxPoolThreadSyncer )
		}

		pool.workingThreadWG.Done()

		fmt.Println( "ATxPool Thread Off: " + ATxPoolThreadSyncer )

	}()


	sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadSyncer] )
	if err != nil {
		return
	}

	for {

		msg, err := sub.Next(subCtx)

		if err != nil {
			return
		}

		select {
		case <- ctx.Done():

			return

		case <- pool.notary.TrustOrNot(msg, core.NotaryMessageChainInfo, pool.cvfs) : {

			info := &AChainInfo.ChainInfo{}

			if err := info.RawMessageDecode(msg.GetData()); err != nil {
				log.Error(err)
				continue
			}

			if info.GenHash != pool.genBlock.GetHash() {
				continue
			}

			for pool.InMining() {
				time.Sleep(time.Millisecond * 100)
			}

			latest, err := pool.cvfs.Indexes().GetLatest()
			if err != nil {
				continue
			}

			pool.syncMutx.Lock()

			if info.LatestBlock.Index == latest.BlockIndex && strings.EqualFold(info.VDBRoot.String(), latest.FullCID.String()) {

				if err := pool.ConfirmBestBlock(info.LatestBlock); err != nil {
					log.Error(err)
					pool.syncMutx.Unlock()
					continue
				}

				// confirm
				if err := pool.cvfs.Indexes().Flush(); err != nil {
					log.Error(err)
				}

			} else if info.LatestBlock.Index <= latest.BlockIndex {
				pool.syncMutx.Unlock()
				continue

			} else {

				if err := pool.cvfs.Indexes().SyncToCID(info.Indexes); err != nil {
					log.Error(err)
				}

			}

			if err := pool.cvfs.SeekToBlock(ABlock.Latest); err != nil {
				log.Warning(err)
			}

			pool.syncMutx.Unlock()
		}
		}
	}

}