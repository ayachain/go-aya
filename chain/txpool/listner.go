package txpool

import (
	"context"
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
)

func (pool *ATxPool) channelListening(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + AtxThreadTopicsListen)

	subscribe, err := pool.ind.PubSub.Subscribe( pool.channelTopics )
	if err != nil {
		return
	}

	tctx, cancel := context.WithCancel(ctx)

	sctx, _ := context.WithCancel(tctx)

	pool.threadClosewg.Add(1)
	
	go func() {
		
		select {
		case <- sctx.Done():

			cancel()

			<- tctx.Done()


		case <- tctx.Done():

			subscribe.Cancel()

			<- sctx.Done()
		}

		fmt.Println("ATxPool Thread Off: " + AtxThreadTopicsListen)

		pool.threadClosewg.Done()
		
	}()

	for {

		msg, err := subscribe.Next(sctx)

		if err != nil {
			return
		}

		rawmsg, err := AKeyStore.BytesToRawMsg(msg.Data)
		if err != nil {
			log.Error(err)
			continue
		}

		if err := pool.rawMessageSwitch(rawmsg); err != nil {
			log.Error(err)
			continue
		}

	}

}