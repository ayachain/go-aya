package txpool

import (
	"context"
	"fmt"
)

func (pool *ATxPool) miningThread(ctx context.Context) {

	for {

		fmt.Println("ATxPool Mining thread power on")
		defer fmt.Println("ATxPool Mining thread power off")

		for {

			select {
			case <-ctx.Done():
				return

			case msg := <- pool.threadChans[AtxThreadsNameMining] :

				ret := <- pool.notary.OnReceiveRawMessage(msg)

				if ret.Err != nil {



				}

			}
		}

	}

}
