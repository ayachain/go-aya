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

	pool.tcmapMutex.Lock()
	pool.threadChans[ATxPoolThreadTxListen] = make(chan []byte, AtxPoolThreadTxListenBuff)
	pool.tcmapMutex.Unlock()

	subCtx, subCancel := context.WithCancel(ctx)

	defer func() {

		subCancel()

		<- subCtx.Done()

		pool.tcmapMutex.Lock()
		cc, exist := pool.threadChans[ATxPoolThreadTxListen]
		if exist {

			close( cc )
			delete(pool.threadChans, ATxPoolThreadTxListen)

		}
		pool.tcmapMutex.Unlock()

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

				pool.threadChans[ATxPoolThreadTxListen] <- msg.Data

			}

		}

	}()


	for {

		select {

		case <- ctx.Done():

			return

		case rawmsg, isOpen := <- pool.threadChans[ATxPoolThreadTxListen]:

			if !isOpen {
				continue
			}

			tx := &ATx.Transaction{}
			if err := tx.RawMessageDecode(rawmsg); err != nil {
				log.Error(err)
				continue
			}

			if err := pool.addRawTransaction(tx); err != nil {

				log.Error(err)
				continue
			}

		}

	}

}