package txpool

import (
	"context"
	MBlock "github.com/ayachain/go-aya/vdb/mblock"
	"github.com/ayachain/go-aya/vdb/node"
	"github.com/prometheus/common/log"
	"time"
)

func (pool *aTxPool) threadMiningBlockRepeater( ctx context.Context ) {

	log.Info("ATxPool Thread On: " + ATxPoolThreadRepeater)
	defer log.Info("ATxPool Thread Off: " + ATxPoolThreadRepeater)

	// subscribe
	sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadRepeater] )
	if err != nil {
		return
	}
	defer sub.Cancel()

	var latestRecvMBlock *MBlock.MBlock
	var latestPublishTime int64

	for {

		msg, err := sub.Next(ctx)
		if err != nil {
			return
		}

		// verify from
		if nd, err := pool.cvfs.Nodes().GetNodeByPeerId(msg.GetFrom().Pretty()); err != nil {

			continue

		} else {

			if nd.Type != node.NodeTypeSuper {

				continue

			}

		}

		mblock := &MBlock.MBlock{}
		if err := mblock.RawMessageDecode(msg.Data); err != nil {
			log.Warn(err)
			continue
		}

		if mblock.GetHash() == latestRecvMBlock.GetHash() {

			if latestPublishTime - time.Now().Unix() < 10 {

				continue

			} else {

				if err := pool.ind.PubSub.Publish(pool.mblockChannel, msg.Data); err != nil {
					log.Error(err)
					continue
				}

				latestPublishTime = time.Now().Unix()

				continue

			}

		} else {

			if err := pool.ind.PubSub.Publish(pool.mblockChannel, msg.Data); err != nil {
				log.Error(err)
				continue
			}

			latestRecvMBlock = mblock
			latestPublishTime = time.Now().Unix()

			continue
		}

	}
}