package txpool

import (
	"context"
	"fmt"
	"github.com/ayachain/go-aya/consensus/core"
	AMsgMinied "github.com/ayachain/go-aya/vdb/minined"
)

func receiptListen(ctx context.Context ) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadReceiptListen)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans.Store(ATxPoolThreadReceiptListen, make(chan []byte, ATxPoolThreadReceiptListenBuff))

	subCtx, subCancel := context.WithCancel(ctx)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans.Load(ATxPoolThreadReceiptListen)
		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete(ATxPoolThreadReceiptListen)
		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadReceiptListen)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadReceiptListen] )

		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				return
			}

			rcp := &AMsgMinied.Minined{}
			if err := rcp.RawMessageDecode( msg.Data ); err == nil {
				log.Infof( "FromPeerID:%v MBlock:%v RetCID:%v", msg.GetFrom().Pretty(), rcp.MBlockHash.String(), rcp.RetCID.String() )
			}

			if <- pool.notary.TrustOrNot(msg, core.NotaryMessageMinedRet, pool.cvfs) {

				cc, _ := pool.threadChans.Load(ATxPoolThreadReceiptListen)

				cc.(chan []byte) <- msg.Data

			}

		}

	}()


	for {

		cc, _ := pool.threadChans.Load(ATxPoolThreadReceiptListen)

		select {
		case <- ctx.Done():

			return

		case rawmsg, isOpen := <- cc.(chan []byte):

			if !isOpen {
				continue
			}

			if rawmsg[0] != AMsgMinied.MessagePrefix {
				continue
			}

			rcp := &AMsgMinied.Minined{}
			err := rcp.RawMessageDecode( rawmsg )
			if err != nil {
				log.Error(err)
				continue
			}

			if pool.miningBlock.GetHash() != rcp.MBlockHash {
				continue
			}

			cblock := pool.miningBlock.Confirm(rcp.RetCID.String())

			if err := pool.doBroadcast(cblock, pool.channelTopics[ATxPoolThreadExecutor] ); err != nil {
				log.Error(err)
				return
			}

		}

	}

}
