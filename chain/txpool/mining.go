package txpool

import (
	"context"
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMsgMined "github.com/ayachain/go-aya/vdb/minined"
	IBlocks "github.com/ipfs/go-block-format"
	"time"
)

func miningThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + AtxThreadMining)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans[AtxThreadMining] = make(chan *AKeyStore.ASignedRawMsg)

	subCtx, subCancel := context.WithCancel(ctx)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans[AtxThreadMining]
		if exist {

			close( cc )
			delete(pool.threadChans, AtxThreadMining)

		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + AtxThreadMining)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[AtxThreadMining] )
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

			pool.threadChans[AtxThreadMining] <- rawmsg
		}

	}()


	for {

		select {

		case <- ctx.Done():
			return

		case msg, isOpen := <- pool.threadChans[AtxThreadMining] :

			stime := time.Now()

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
				return
			}

			group, err := pool.notary.MiningBlock(mblock, cVFS)
			if err != nil {

				if err := cVFS.Close(); err != nil {
					log.Error(err)
					return
				}

				continue
			}

			if err := cVFS.Close(); err != nil {
				log.Error(err)
				return
			}

			groupbytes := group.Encode()
			gblock := IBlocks.NewBlock(groupbytes)
			if err := pool.ind.Blocks.AddBlock(gblock); err != nil {
				log.Error(err)
				return
			}

			mRet := &AMsgMined.Minined{
				MBlockHash:mblock.GetHash(),
				RetCID:gblock.Cid(),
			}

			if err := pool.doBroadcast(mRet, pool.channelTopics[AtxThreadReceiptListen]); err != nil {
				log.Error(err)
				return
			}


			fmt.Println("AtxThreadMining HandleTime:", time.Since(stime))
		}
	}

}
