package txpool

import (
	"context"
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"time"
)

func txListenThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + AtxThreadTxListen)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans[AtxThreadTxListen] = make(chan *AKeyStore.ASignedRawMsg)

	subCtx, subCancel := context.WithCancel(ctx)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans[AtxThreadTxListen]
		if exist {

			close( cc )
			delete(pool.threadChans, AtxThreadTxListen)

		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + AtxThreadTxListen)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[AtxThreadTxListen] )

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

			pool.threadChans[AtxThreadTxListen] <- rawmsg
		}

	}()


	for {

		select {

		case <- ctx.Done():

			return

		case msg, isOpen := <- pool.threadChans[AtxThreadTxListen]:

			stime := time.Now()

			if !isOpen {
				continue
			}

			if err := pool.addRawTransaction(msg); err != nil {

				log.Error(err)
				continue
			}

			fmt.Println("AtxThreadTxListen HandleTime:", time.Since(stime))
		}

	}

}