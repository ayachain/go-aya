package txpool

import (
	"context"
	"fmt"
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


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadChainInfo] )
		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				return
			}

			info := &AChainInfo.ChainInfo{}

			if err := info.RawMessageDecode(msg.GetData()); err != nil {
				log.Error(err)
				continue
			}

			if <- pool.notary.ShouldDoSync(info, pool.cvfs) {

				// need sync
				pool.syncMutx.Lock()
				pool.syncChainInfo = info

				getctx, cancel := context.WithCancel( context.TODO() )
				pool.syncCancel = cancel
				pool.ind.Pinning.PinWithMode(info.Indexes, pin.Any)

				adbipath := AIndexes.AIndexesKeyPathPrefix + pool.genBlock.ChainID
				dsk := datastore.NewKey(adbipath)

				if _, err := pool.ind.DAG.Get( getctx, info.Indexes ); err != nil {
					log.Warning("syncing failed in %d waiting next chain info.", info.LatestBlock.Index)
					goto loopBreakByErr
				}

				if err := pool.ind.Repo.Datastore().Put(dsk, info.Indexes.Bytes()); err != nil {
					log.Warning(err)
					goto loopBreakByErr
				}

				// restart cvfs
				if err := pool.cvfs.Restart( info.VDBRoot ); err != nil {
					log.Warning(err)
					goto loopBreakByErr
				}

				loopBreakByErr:

					pool.syncCancel = nil
					pool.syncChainInfo = nil
					pool.syncMutx.Unlock()

					continue
			}

		}

	}()




}