package txpool

import (
	"context"
	"fmt"
	"github.com/ayachain/go-aya/consensus/core"
	AChainInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	AIndexes "github.com/ayachain/go-aya/vdb/indexes"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/pin"
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

		if <- pool.notary.TrustOrNot(msg, core.NotaryMessageChainInfo, pool.cvfs ) {

			info := &AChainInfo.ChainInfo{}

			if err := info.RawMessageDecode(msg.GetData()); err != nil {
				log.Error(err)
				continue
			}

			latest, err := pool.cvfs.Indexes().GetLatest()
			if err != nil {
				continue
			}

			if info.LatestBlock.Index <= latest.BlockIndex {
				continue
			}

			// need sync
			pool.syncMutx.Lock()
			pool.ind.Pinning.PinWithMode(info.Indexes, pin.Any)

			adbipath := AIndexes.AIndexesKeyPathPrefix + pool.genBlock.ChainID
			dsk := datastore.NewKey(adbipath)

			if err := pool.ind.Repo.Datastore().Put(dsk, info.Indexes.Bytes()); err != nil {
				log.Warning(err)
				goto loopBreakByErr
			}

			// restart cvfs
			if err := pool.cvfs.Restart( info.VDBRoot ); err != nil {
				log.Warning(err)
				goto loopBreakByErr
			}

			log.Infof("SyncBlockIndex %08d:%v, ", info.LatestBlock.Index, info.VDBRoot.String() )

			loopBreakByErr:

				pool.syncMutx.Unlock()

				continue
		}

	}

}