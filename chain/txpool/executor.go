package txpool

import (
	"context"
	"fmt"
	ATaskGroup "github.com/ayachain/go-aya/consensus/core/worker"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	AMsgBlock "github.com/ayachain/go-aya/vdb/block"
	AChainInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	"github.com/ipfs/go-cid"
	"time"
)

func blockExecutorThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + AtxThreadExecutor)

	pool := ctx.Value("Pool").(*ATxPool)

	subCtx, subCancel := context.WithCancel(ctx)

	pool.threadChans[AtxThreadExecutor] = make(chan *AKeyStore.ASignedRawMsg)

	pool.workingThreadWG.Add(1)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans[AtxThreadExecutor]
		if exist {

			close( cc )
			delete(pool.threadChans, AtxThreadExecutor)

		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + AtxThreadExecutor)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[AtxThreadExecutor] )
		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				return
			}

			rawmsg, err := AKeyStore.BytesToRawMsg(msg.Data)
			if err != nil {
				log.Error(err)
				continue
			}

			pool.threadChans[AtxThreadExecutor] <- rawmsg
		}

	}()


	for {

		select {
		case <- ctx.Done():

			return

		case msg, isOpen := <- pool.threadChans[AtxThreadExecutor] :

			stime := time.Now()

			if !isOpen {
				continue
			}

			if !msg.Verify() {
				continue
			}

			cblock := &AMsgBlock.Block{}
			if err := cblock.RawMessageDecode(msg.Content); err != nil {
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

				if err := pool.doBroadcast(info, pool.channelTopics[AtxThreadChainInfo]); err != nil {
					log.Error(err)
					return
				}

			} else {
				return
			}


			_ = pool.UpdateBestBlock(cblock)

			pool.miningBlock = nil
			
			pool.DoPackMBlock()

			fmt.Printf("Confrim Block %06d:\tCID:%v\n", cblock.Index, latestCid.String())

			fmt.Println("AtxThreadExecutor HandleTime:", time.Since(stime))
		}
	}
}