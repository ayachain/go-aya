package common

import (
	"context"
	"errors"
	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/miekg/dns"
)

var(
	ErrNotRunningThread = errors.New("thread not running")
)

type Thread struct {

	consumer Consumer
	producer Producer

	insleep bool
	sig chan AThreadSemaphore
	Type AThreadRoleType
}

func (trd *Thread) Semaphore( sig AThreadSemaphore ) {
	trd.sig <- sig
}

func (trd *Thread)  Start(ctx context.Context, role AThreadRoleType, subscription pubsub.Subscription) {

	messagepool := make(chan *pubsub.Message, 128)

	subctx, subcancel := context.WithCancel(context.Background())

	go func() {

		for {

			msg, err := subscription.Next(subctx)
			if err != nil {
				return
			}

			messagepool <- msg
		}

	}()

	go func() {

		for {

				select {

				case <- ctx.Done():
					subcancel()
					<- subctx.Done()
					return

				case s := <- trd.sig:
					switch s {
					case ThreadSemaphoreResume: trd.insleep = false; continue
					case ThreadSemaphoreSleep: trd.insleep = true; continue
					case ThreadSemaphoreStop: return
					}

				default:

					if trd.insleep {
						continue
					} else {

					}

				}
		}

	}()

}

func (trd *Thread) Wakeup() {
	trd.wakeup <- nil
}

func (trd *Thread) ChangeRole(roleType ThreadRoleType) {

	trd.Type = roleType

	trd.sig <- nil

}

func (trd *Thread)  Shutdown() error {
	return nil
}