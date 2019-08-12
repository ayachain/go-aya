package txpool

import (
	"context"
	"fmt"
	"github.com/ayachain/go-aya/consensus/core"
	"github.com/ayachain/go-aya/vdb/block"
	AChainInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	"strings"
)

func syncListener(ctx context.Context ) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadChainInfo)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans.Store(ATxPoolThreadChainInfo, make(chan []byte, ATxPoolThreadChainInfoBuff))

	subCtx, subCancel := context.WithCancel(ctx)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans.Load( ATxPoolThreadChainInfo )
		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete( ATxPoolThreadChainInfo )
		}

		pool.workingThreadWG.Done()

		fmt.Println( "ATxPool Thread Off: " + ATxPoolThreadChainInfo )

	}()


	sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadChainInfo] )
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

			latest, err := pool.cvfs.Indexes().GetLatest()
			if err != nil {
				continue
			}

			if info.LatestBlock.Index == latest.BlockIndex && strings.EqualFold(info.VDBRoot.String(), latest.FullCID.String()) {

				// confirm
				if err := pool.cvfs.Indexes().Flush(); err != nil {
					log.Error(err)
					continue
				}

				if err := pool.ConfirmBestBlock(info.LatestBlock); err != nil {
					log.Error(err)
				}

			} else if info.LatestBlock.Index <= latest.BlockIndex {

				continue

			} else {

				// need sync
				pool.syncMutx.Lock()

				if err := pool.cvfs.Indexes().SyncToCID(info.Indexes); err != nil {
					log.Error(err)
					goto loopBreakByErr
				}

			}

			// seek block to cvfs
			if err := pool.cvfs.SeekToBlock(block.Latest); err != nil {
				log.Warning(err)
				goto loopBreakByErr
			}

		loopBreakByErr:

			pool.syncMutx.Unlock()

			continue
		}
		}
	}

}