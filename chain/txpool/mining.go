package txpool

import (
	"context"
	"fmt"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMsgMined "github.com/ayachain/go-aya/vdb/minined"
	IBlocks "github.com/ipfs/go-block-format"
)

func (pool *ATxPool) miningThread(ctx context.Context) {

	for {

		fmt.Println("ATxPool Mining thread power on")
		defer fmt.Println("ATxPool Mining thread power off")

		for {

			select {

			case <-ctx.Done():
				return

			case msg := <- pool.threadChans[AtxThreadsNameMining] :

				if !msg.Verify() {
					break
				}

				mblock := &AMsgMBlock.MBlock{}

				if err := mblock.RawMessageDecode(msg.Content); err != nil {
					break
				}

				pool.miningBlock = mblock

				group, err := pool.notary.MiningBlock(mblock)
				if err != nil {
					break
				}

				groupbytes := group.Encode()
				gblock := IBlocks.NewBlock(groupbytes)
				if err := pool.ind.Blocks.AddBlock(gblock); err != nil {
					break
				}

				mRet := &AMsgMined.Minined{
					MBlockHash:mblock.GetHash(),
					RetCID:gblock.Cid(),
				}

				if err := pool.DoBroadcast(mRet); err != nil {
					pool.PowerOff(err)
				}

				//c, exist := pool.threadChans[AtxThreadsNameReceiptListen]
				//if exist {
				//
				//	signmsg, err := pool.sign(mRet)
				//	if err != nil {
				//		break
				//	}
				//
				//	c <- signmsg
				//}

			}
		}

	}

}
