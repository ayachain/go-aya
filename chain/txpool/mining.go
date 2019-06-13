package txpool

import (
	"context"
	"fmt"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMsgMined "github.com/ayachain/go-aya/vdb/minined"
	IBlocks "github.com/ipfs/go-block-format"
)

func (pool *ATxPool) miningThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On : " + AtxThreadMining)
	defer fmt.Println("ATxPool Thread Off : " + AtxThreadMining)

	for {

		select {

		case <-ctx.Done():
			return

		case msg := <- pool.threadChans[AtxThreadMining] :

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
				return
			}

		}
	}

}
