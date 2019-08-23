package txpool

import (
	"context"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/common/log"
	"sync"
)

func (pool *aTxPool) threadMiningBlockRepeater( ctx context.Context, awaiter *sync.WaitGroup ) {

	awaiter.Add(1)
	defer awaiter.Done()

	// subscribe
	sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadRepeater] )
	if err != nil {
		return
	}
	defer sub.Cancel()

	for {

		msg, err := sub.Next(ctx)
		if err != nil {
			return
		}

		// verify from
		if nd, err := pool.cvfs.Nodes().GetNodeByPeerId(msg.GetFrom().Pretty()); err != nil {

			continue

		} else {

			if nd.Type != im.NodeType_Super {
				continue
			}

		}

		mblock := &im.Block{}
		if err := proto.Unmarshal(msg.Data, mblock); err != nil {
			log.Warn(err)
			continue
		}

		if err := pool.ind.PubSub.Publish(pool.mblockChannel, msg.Data); err != nil {
			log.Error(err)
			continue
		}

		pool.lmblock = mblock

		continue
	}
}