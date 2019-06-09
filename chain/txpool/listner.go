package txpool

import (
	"context"
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
)

func (pool *ATxPool) channelListening(ctx context.Context) {

	fmt.Println("ATxPool channel listening thread power on")
	defer fmt.Println("ATxPool channel listening thread power off")

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

		if err := pool.RawMessageSwitch(rawmsg); err != nil {
			continue
		}

	}

}