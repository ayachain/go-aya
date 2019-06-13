package txpool

import (
	"context"
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
)

func (pool *ATxPool) channelListening(ctx context.Context) {

	fmt.Println("ATxPool Thread On : " + AtxThreadTopicsListen)
	defer fmt.Println("ATxPool Thread Off : " + AtxThreadTopicsListen)

	subscribe, err := pool.ind.PubSub.Subscribe( pool.channelTopics )
	if err != nil {
		return
	}

	for {

		msg, err := subscribe.Next(ctx)

		if err != nil {
			return
		}

		rawmsg, err := AKeyStore.BytesToRawMsg(msg.Data)
		if err != nil {
			continue
		}

		if err := pool.rawMessageSwitch(rawmsg); err != nil {
			continue
		}

	}

}