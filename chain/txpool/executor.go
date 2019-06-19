package txpool

import (
	"context"
	"fmt"
	ATaskGroup "github.com/ayachain/go-aya/consensus/core/worker"
	AMsgBlock "github.com/ayachain/go-aya/vdb/block"
	AChainInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	"github.com/ipfs/go-cid"
)

func (pool *ATxPool) blockExecutorThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + AtxThreadExecutor)

	tctx, cancel := context.WithCancel(ctx)

	pool.threadClosewg.Add(1)

	for {

		select {
		case <- tctx.Done():

			close(pool.threadChans[AtxThreadExecutor])

			delete(pool.threadChans, AtxThreadExecutor)

			fmt.Println("ATxPool Thread Off: " + AtxThreadExecutor)

			pool.threadClosewg.Done()

			return

		case msg, isOpen := <- pool.threadChans[AtxThreadExecutor] :

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


			batchBlock, err := pool.ind.Blocks.GetBlock( tctx, bcid )
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
				cancel()
				continue
			}

			if err := pool.cvfs.Indexes().PutIndexBy( cblock.Index, cblock.GetHash(), latestCid ); err != nil {
				log.Error(err)
				cancel()
				continue
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

				if err := pool.DoBroadcast(info); err != nil {
					log.Error(err)
					cancel()
					continue
				}

			} else {

				cancel()
				continue
			}

			_ = pool.UpdateBestBlock(cblock)

			pool.DoPackMBlock()

			fmt.Printf("Confrim Block %06d:\tCID:%v\n", cblock.Index, latestCid.String())
		}
	}
}