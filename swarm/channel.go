package swarm

import (
	"context"
	iCore "github.com/ayachain/go-aya/ipfsapi"
	iFace "github.com/ipfs/interface-go-ipfs-core"
)

type channel struct {

	topics string

	broadcastCancer context.CancelFunc

	listenCancer context.CancelFunc
	listenScription iFace.PubSubSubscription
	listenHandle func(msg iFace.PubSubMessage)
}

func NewChannel( aappns string, handle func(msg iFace.PubSubMessage) ) *channel  {
	return &channel{
		topics:aappns,
		listenHandle:handle,
	}
}

func (c *channel) DoBroadcast(ct []byte) error {

	var ctx context.Context

	ctx, c.broadcastCancer = context.WithCancel( context.Background() )

	return iCore.IAPI.PubSub().Publish(ctx, c.topics, ct)
}

func (c *channel) Close() error  {
	return c.listenScription.Close()
}

func (c *channel) Open() error {

	var ctx context.Context
	var err error

	ctx, c.listenCancer = context.WithCancel( context.Background() )
	c.listenScription, err = iCore.IAPI.PubSub().Subscribe( ctx, c.topics )

	if err != nil {
		return err
	}

	go func() {
		for {
			msg, err := c.listenScription.Next(ctx)
			if err != nil {
				c.listenCancer()
				return
			} else {
				c.listenHandle(msg)
			}
		}
	}()

}