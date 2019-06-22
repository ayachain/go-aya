package txpool

import (
	"context"
	"fmt"
	"github.com/ayachain/go-aya/consensus/core"
	AMsgMinied "github.com/ayachain/go-aya/vdb/minined"
	"github.com/syndtr/goleveldb/leveldb"
)

func receiptListen(ctx context.Context ) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadReceiptListen)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans.Store(ATxPoolThreadReceiptListen, make(chan []byte, ATxPoolThreadReceiptListenBuff))

	subCtx, subCancel := context.WithCancel(ctx)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans.Load(ATxPoolThreadReceiptListen)
		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete(ATxPoolThreadReceiptListen)
		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadReceiptListen)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadReceiptListen] )

		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				return
			}

			if <- pool.notary.TrustOrNot(msg, core.NotaryMessageMinedRet, pool.cvfs) {

				cc, _ := pool.threadChans.Load(ATxPoolThreadReceiptListen)

				cc.(chan []byte) <- msg.Data

			}

		}

	}()


	for {

		cc, _ := pool.threadChans.Load(ATxPoolThreadReceiptListen)

		select {
		case <- ctx.Done():

			return

		case rawmsg, isOpen := <- cc.(chan []byte):

			if !isOpen {
				continue
			}

			if rawmsg[0] != AMsgMinied.MessagePrefix {
				continue
			}

			rcp := &AMsgMinied.Minined{}
			err := rcp.RawMessageDecode( rawmsg )
			if err != nil {
				log.Error(err)
				continue
			}

			if pool.miningBlock.GetHash() != rcp.MBlockHash {
				continue
			}

			rcidstr := rcp.RetCID.String()

			receiptKey := []byte(TxReceiptPrefix)

			copy( receiptKey[1:], rcp.MBlockHash.Bytes() )

			if pool.workmode == AtxPoolWorkModeOblivioned {

				batch := &leveldb.Batch{}
				batch.Delete(receiptKey)
				batch.Delete(rcp.MBlockHash.Bytes())

				if err := pool.storage.Write(batch, nil); err != nil {
					log.Error(err)
					return
				}


				cblock := pool.miningBlock.Confirm(rcidstr)

				if err := pool.doBroadcast(cblock, pool.channelTopics[ATxPoolThreadExecutor] ); err != nil {
					log.Error(err)
					return
				}

			}

		}

	}

}
