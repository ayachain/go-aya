package common

import (
	"context"
	"errors"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ipfs/go-ipfs/core"
)

var(
	ErrNotRunningThread = errors.New("thread not running")
)


type Thread interface {

	Consumer

	Producer

	Start(ctx context.Context, ind *core.IpfsNode, topic string)

	Semaphore( sig AThreadSemaphore )
}


type AThread struct {

	Thread

	sig chan AThreadSemaphore

	Type AThreadRoleType

	Topics string
}

func (trd *AThread) Start(ctx context.Context, ind *core.IpfsNode, topic string) {

	trd.Topics = topic

	funChangeChan := make(chan func() <- chan struct{})
	var cancel context.CancelFunc

	go func() {

		changeSig := make(chan struct{})

		for {

			select {

			case <- ctx.Done():

				if cancel != nil {
					cancel()
				}
				return

			case s := <- trd.sig:

				switch s {

				case ThreadSemaphoreStop:

					funChangeChan <- nil

					cancel()

					return

				case ThreadSemaphoreConsumer:

					if trd.Type != ThreadRoleTypeConsumer {
						trd.Type = ThreadRoleTypeConsumer
						changeSig <- nil
					}

					continue


				case ThreadSemaphoreProducer:

					if trd.Type == ThreadRoleTypeProcucer {
						trd.Type = ThreadRoleTypeProcucer
						changeSig <- nil
					}

					continue

				}

			case <- changeSig:

				if cancel != nil {
					cancel()
				}

				switch trd.Type {
				case ThreadRoleTypeConsumer:

					cel, fn := trd.newConsumerDaemonThread(ind)
					cancel = cel
					funChangeChan <- fn

				case ThreadRoleTypeProcucer:

					cel, fn := trd.newProducerDaemonThread(ind)
					cancel = cel
					funChangeChan <- fn
				}

			}
		}

	}()

	go func() {

		for {

			select {
			case fun := <- funChangeChan:

				if fun == nil {
					return
				}

				<- fun()
			}

		}

	}()

}

func (trd *AThread) Semaphore( sig AThreadSemaphore ) {
	trd.sig <- sig
}

func (trd *AThread) newProducerDaemonThread( ind *core.IpfsNode ) (context.CancelFunc, func() <- chan struct{} ) {

	// Producer thread daemon
	producerCtx, producerCancel := context.WithCancel(context.Background())
	producerWorkThread := func() <- chan struct{} {

		doneChan := make(chan struct{})

		go func() {

			for {

				msg, err := trd.DoProduce(producerCtx)

				if err != nil {
					doneChan <- nil
					return
				}

				rawmsg, err := msg.Bytes()
				if err != nil {
					log.Error(err)
					continue
				}

				if err := ind.PubSub.Publish(trd.Topics, rawmsg); err != nil {
					log.Error(err)
					continue
				}

			}

		}()

		return doneChan
	}

	return producerCancel, producerWorkThread
}

func (trd *AThread) newConsumerDaemonThread( ind *core.IpfsNode ) (context.CancelFunc, func() <- chan struct{} ) {

	// Consumer thread daemon
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	consumerWorkThread := func() <- chan struct{} {

		doneChan := make(chan struct{})

		subscription, err := ind.PubSub.Subscribe(trd.Topics)
		if err != nil {
			doneChan <- nil
			return doneChan
		}

		go func() {

			for {

				msg, err := subscription.Next(consumerCtx)

				if err != nil {
					doneChan <- nil
					return
				}

				rawmsg, err := AKeyStore.BytesToRawMsg(msg.Data)
				if err != nil {
					log.Error(err)
					continue
				}

				<- trd.DoConsume(rawmsg)

				continue

			}

		}()

		return doneChan
	}

	return consumerCancel, consumerWorkThread
}