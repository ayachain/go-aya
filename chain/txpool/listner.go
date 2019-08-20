package txpool

import (
	"context"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/prometheus/common/log"
)

func (pool *aTxPool) threadTransactionListener( ctx context.Context ) {

	//log.Info("ATxPool Thread On: " + ATxPoolThreadTxListen)
	//defer log.Info("ATxPool Thread Off: " + ATxPoolThreadTxListen)

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

		tx := &ATx.Transaction{}
		if err := tx.RawMessageDecode(msg.Data); err != nil {
			log.Warn(err)
			continue
		}

		if err := pool.storeTransaction(tx); err != nil {
			log.Warn(err)
			continue
		}
	}
}