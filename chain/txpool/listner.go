package txpool

import (
	"context"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/common/log"
	"sync"
)

func (pool *aTxPool) threadTransactionListener( ctx context.Context, awaiter *sync.WaitGroup ) {

	awaiter.Add(1)
	defer awaiter.Done()

	// subscribe
	sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadTxListen] )
	if err != nil {
		return
	}
	defer sub.Cancel()

	for {

		msg, err := sub.Next(ctx)
		if err != nil {
			return
		}

		tx := &im.Transaction{}

		if err := proto.Unmarshal(msg.Data, tx); err != nil {
			log.Warn(err)
			continue
		}

		if err := pool.storeTransaction(tx); err != nil {
			log.Warn(err)
			continue
		}
	}
}