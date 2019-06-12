package txpool

import (
	"context"
	"fmt"
	ATaskGroup "github.com/ayachain/go-aya/consensus/core/worker"
	AMsgBlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ipfs/go-cid"
	"time"
)

func (pool *ATxPool) blockPackageThread(ctx context.Context) {

	fmt.Println("ATxPool block package thread power on")
	defer fmt.Println("ATxPool block package thread power off")

	for {

		select {
		case <-ctx.Done():
			return

		case msg := <- pool.threadChans[AtxThreadsNameBlockPacage] :

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


				putBlockTransaction, err := pool.cvfs.OpenTransaction()
				if err != nil {
					fmt.Println(err)
					break
				}


				if err := putBlockTransaction.Write(group); err != nil {
					fmt.Println(err)
					break
				}


				if err := putBlockTransaction.Commit(); err != nil {
					fmt.Println(err)
					break
				}


				if err := pool.cvfs.Blocks().AppendBlocks( cblock ); err != nil {
					fmt.Println(err)
					break
				}

				cctx, ccancel := context.WithTimeout(context.TODO(), time.Second  * 60)
				defer ccancel()


				fllcid, err := pool.cvfs.Flush(cctx)
				if err != nil {
					fmt.Println(err)
					break
				}

				if err := pool.cvfs.Indexes().PutIndexBy( cblock.Index, cblock.GetHash(), fllcid ); err != nil {
					fmt.Println(err)
					break
				}

				if err := pool.UpdateBestBlock(); err != nil {

					fmt.Println("UpdateBestBlockErr:" + err.Error())

				} else {

					fmt.Printf( "ConfirmBlock:%v FullCID:%v\n", cblock.Index, fllcid)

				}

			}

		}
	}

}