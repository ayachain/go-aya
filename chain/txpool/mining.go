package txpool

import (
	"context"
	"fmt"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMsgMined "github.com/ayachain/go-aya/vdb/minined"
	IBlocks "github.com/ipfs/go-block-format"
)

func (pool *ATxPool) miningThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + AtxThreadMining)

	tctx, cancel := context.WithCancel(ctx)

	pool.threadClosewg.Add(1)

	for {

		select {

		case <-tctx.Done():

			fmt.Println("ATxPool Thread Off: " + AtxThreadMining)

			pool.threadClosewg.Done()

			return

		case msg, isOpen := <- pool.threadChans[AtxThreadMining] :

			if !isOpen {
				continue
			}

			if !msg.Verify() {
				continue
			}

			mblock := &AMsgMBlock.MBlock{}
			if err := mblock.RawMessageDecode(msg.Content); err != nil {
				log.Error(err)
				continue
			}

			cVFS, err := pool.cvfs.NewCVFSCache()
			if err != nil {
				log.Error(err)
				cancel()
				continue
			}

			group, err := pool.notary.MiningBlock(mblock, cVFS)
			if err != nil {

				if err := cVFS.Close(); err != nil {
					log.Error(err)
					cancel()
				}

				continue
			}

			if err := cVFS.Close(); err != nil {
				log.Error(err)
				cancel()
				continue
			}

			groupbytes := group.Encode()
			gblock := IBlocks.NewBlock(groupbytes)
			if err := pool.ind.Blocks.AddBlock(gblock); err != nil {
				log.Error(err)
				cancel()
				continue
			}

			mRet := &AMsgMined.Minined{
				MBlockHash:mblock.GetHash(),
				RetCID:gblock.Cid(),
			}

			if err := pool.DoBroadcast(mRet); err != nil {
				log.Error(err)
				cancel()
				continue
			}

		}
	}

}
