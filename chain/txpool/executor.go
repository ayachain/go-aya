package txpool

import (
	"context"
	"fmt"
	"github.com/ayachain/go-aya/consensus/core"
	ATaskGroup "github.com/ayachain/go-aya/consensus/core/worker"
	AMsgBlock "github.com/ayachain/go-aya/vdb/block"
	AChainInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	AElectoral "github.com/ayachain/go-aya/vdb/electoral"
	"github.com/ipfs/go-cid"
)

func blockExecutorThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadExecutor)

	pool := ctx.Value("Pool").(*ATxPool)

	subCtx, subCancel := context.WithCancel(ctx)

	pool.threadChans.Store(ATxPoolThreadExecutor, make(chan []byte, ATxPoolThreadExecutorBuff) )

	pool.workingThreadWG.Add(1)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans.Load(ATxPoolThreadExecutor)
		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete(ATxPoolThreadExecutor)
		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadExecutor)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadExecutor] )

		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				return
			}

			if <- pool.notary.TrustOrNot(msg, core.NotaryMessageConfirmBlock, pool.cvfs) {

				cc, _ := pool.threadChans.Load(ATxPoolThreadExecutor)

				cc.(chan []byte) <- msg.Data

			}
		}

	}()


	for {

		cc, _ := pool.threadChans.Load(ATxPoolThreadExecutor)

		select {
		case <- ctx.Done():

			return

		case rawmsg, isOpen := <- cc.(chan []byte):

			if !isOpen {
				continue
			}

			cblock := &AMsgBlock.Block{}
			if err := cblock.RawMessageDecode(rawmsg); err != nil {
				log.Error(err)
				continue
			}

			bcid, err := cid.Decode(cblock.ExtraData)
			if err != nil {
				log.Error(err)
				continue
			}


			batchBlock, err := pool.ind.Blocks.GetBlock( context.TODO(), bcid )
			if err != nil {
				log.Error(err)
				continue
			}


			group := ATaskGroup.NewGroup()
			if err := group.Decode(batchBlock.RawData()); err != nil {
				log.Error(err)
				continue
			}

			// Append Block
			group.Put(AMsgBlock.DBPath, cblock.GetHash().Bytes(), cblock.Encode() )

			latestCid, err := pool.cvfs.WriteTaskGroup(group)
			if err != nil {
				log.Error(err)
				return
			}

			if err := pool.cvfs.Indexes().PutIndexBy( cblock.Index, cblock.GetHash(), latestCid ); err != nil {
				log.Error(err)
				return
			}

			indexCid := pool.cvfs.Indexes().Flush()

			if !indexCid.Equals(cid.Undef) {

				// broadcast chain info
				info := &AChainInfo.ChainInfo {
					GenHash:pool.genBlock.GetHash(),
					VDBRoot:latestCid,
					LatestBlock:cblock,
					Indexes:indexCid,
				}

				if err := pool.doBroadcast(info, pool.channelTopics[ATxPoolThreadChainInfo]); err != nil {
					log.Error(err)
					return
				}

			} else {
				return
			}

			_ = pool.UpdateBestBlock(cblock)

			pool.miningBlock = nil

			if pool.packerState == AElectoral.ATxPackStateNextMaster {
				pool.packerState = AElectoral.ATxPackStateMaster
				pool.DoPackMBlock()
			}

			pool.notary.NewBlockHasConfirm()

			log.Infof("Confrim Block %08d:%v", cblock.Index, latestCid.String())
		}
	}
}