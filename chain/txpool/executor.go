package txpool

import (
	"context"
	"fmt"
	ATaskGroup "github.com/ayachain/go-aya/consensus/core/worker"
	AMsgBlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ipfs/go-cid"
	"time"
)

func (pool *ATxPool) blockExecutorThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On : " + AtxThreadExecutor)
	defer fmt.Println("ATxPool Thread Off : " + AtxThreadExecutor)

	for {

		select {
		case <-ctx.Done():
			return

		case msg := <- pool.threadChans[AtxThreadExecutor] :

			if pool.miningBlock != nil {

				if !msg.Verify() {
					break
				}

				cblock := &AMsgBlock.Block{}
				if err := cblock.RawMessageDecode(msg.Content); err != nil {
					break
				}


				bcid, err := cid.Decode(cblock.ExtraData)
				if err != nil {
					break
				}


				bctx, cancel := context.WithTimeout(context.TODO(), time.Second  * 60)
				defer cancel()


				batchBlock, err := pool.ind.Blocks.GetBlock( bctx, bcid )
				if err != nil {
					fmt.Println(err)
					break
				}


				group := ATaskGroup.NewGroup()
				if err := group.Decode(batchBlock.RawData()); err != nil {
					fmt.Println(err)
					break
				}

				// Append Block
				if err := pool.cvfs.Blocks().AppendBlocks(group, cblock ); err != nil {
					fmt.Println(err)
					break
				}

				latestCid, err := pool.cvfs.WriteTaskGroup(group)
				if err != nil {
					fmt.Println(err)
					break
				}

				prevBlockIdx, err := pool.cvfs.Indexes().GetIndex(cblock.Index - 1)
				if err != nil {
					fmt.Println(err)
					break
				}
				fmt.Printf("NewBlock:%06d\t%v\tPrev:%v\n", cblock.Index, latestCid.String(), prevBlockIdx.FullCID )

				if err := pool.cvfs.Indexes().PutIndexBy( cblock.Index, cblock.GetHash(), latestCid ); err != nil {
					fmt.Println(err)
					break
				}

				pool.UpdateBestBlock()
			}

		}
	}
}