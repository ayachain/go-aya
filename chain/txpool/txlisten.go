package txpool

import (
	"context"
	"fmt"
	"github.com/ayachain/go-aya/consensus/core"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
)

func txListenThread(ctx context.Context ) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadTxListen)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans.Store(ATxPoolThreadTxListen, make(chan []byte, AtxPoolThreadTxListenBuff))

	subCtx, subCancel := context.WithCancel(ctx)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans.Load(ATxPoolThreadTxListen)
		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete(ATxPoolThreadTxListen)

		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadTxListen)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadTxListen] )

		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				return
			}

			if <- pool.notary.TrustOrNot(msg, core.NotaryMessageTransaction, pool.cvfs) {

				cc, _ := pool.threadChans.Load(ATxPoolThreadTxListen)

				cc.(chan []byte) <- msg.Data

			}

		}

	}()


	for {

		cc, _ := pool.threadChans.Load(ATxPoolThreadTxListen)

		select {

		case <- ctx.Done():

			return

		case rawmsg, isOpen := <- cc.(chan []byte):

			if !isOpen {
				continue
			}

			tx := &ATx.Transaction{}
			if err := tx.RawMessageDecode(rawmsg); err != nil {
				log.Error(err)
				continue
			}

			if err := pool.PushTransaction(tx); err != nil {
				log.Error(err)
				continue
			}

		}

	}

}