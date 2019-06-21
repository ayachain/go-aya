package txpool

import (
	"context"
	"encoding/json"
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	AMsgMinied "github.com/ayachain/go-aya/vdb/minined"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
)

func receiptListen(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + AtxThreadReceiptListen)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans[AtxThreadReceiptListen] = make(chan *AKeyStore.ASignedRawMsg)

	subCtx, subCancel := context.WithCancel(ctx)


	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans[AtxThreadReceiptListen]
		if exist {

			close( cc )
			delete(pool.threadChans, AtxThreadReceiptListen)

		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + AtxThreadReceiptListen)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[AtxThreadReceiptListen] )
		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				return
			}

			rawmsg, err := AKeyStore.BytesToRawMsg(msg.Data)
			if err != nil {
				log.Error(err)
				continue
			}

			pool.threadChans[AtxThreadReceiptListen] <- rawmsg

		}

	}()


	for {

		select {
		case <- ctx.Done():

			return

		case rmsg, isOpen := <- pool.threadChans[AtxThreadReceiptListen] :

			stime := time.Now()

			if !isOpen {
				continue
			}

			if rmsg.Content[0] != AMsgMinied.MessagePrefix {
				continue
			}

			from, err := rmsg.ECRecover()
			if err != nil {
				log.Error(err)
				continue
			}

			fromAst, err := pool.cvfs.Assetses().AssetsOf(*from)
			if err != nil {
				log.Error(err)
				continue
			}

			rcp := &AMsgMinied.Minined{}
			err = rcp.RawMessageDecode( rmsg.Content )
			if err != nil {
				log.Error(err)
				continue
			}

			if pool.miningBlock.GetHash() != rcp.MBlockHash {
				continue
			}

			receiptKey := []byte(TxReceiptPrefix)
			copy( receiptKey[1:], rcp.MBlockHash.Bytes() )

			exist, err := pool.storage.Has(receiptKey, nil)
			if err != nil {
				log.Error(ErrStorageLowAPIExpected)
				return
			}

			rcidstr := rcp.RetCID.String()
			receiptMap := make(map[string]uint64)
			if exist {

				value, err := pool.storage.Get(receiptKey, nil)
				if err != nil {
					log.Error(ErrStorageLowAPIExpected)
					return
				}

				if json.Unmarshal(value, receiptMap) != nil {
					log.Error(ErrStorageRawDataDecodeExpected)
					return
				}

				ocount, vexist := receiptMap[rcidstr]
				if vexist {
					receiptMap[rcidstr] = ocount + fromAst.Vote
				} else {
					receiptMap[rcidstr] = fromAst.Vote
				}

			} else {
				receiptMap[rcidstr] = fromAst.Vote
			}


			rmapbs, err := json.Marshal(receiptMap)
			if err != nil {
				log.Error(ErrStorageLowAPIExpected)
				return
			}

			if err := pool.storage.Put( receiptKey, rmapbs, nil ); err != nil {
				log.Error(ErrStorageLowAPIExpected)
				return
			}


			if receiptMap[rcidstr] > pool.ownerAsset.Vote * 3 || pool.workmode == AtxPoolWorkModeOblivioned {

				batch := &leveldb.Batch{}
				batch.Delete(receiptKey)
				batch.Delete(rcp.MBlockHash.Bytes())

				if err := pool.storage.Write(batch, nil); err != nil {
					log.Error(err)
					return
				}


				cblock := pool.miningBlock.Confirm(rcidstr)

				if err := pool.doBroadcast(cblock, pool.channelTopics[AtxThreadExecutor] ); err != nil {
					log.Error(err)
					return
				}

			}

			fmt.Println("AtxThreadReceiptListen HandleTime:", time.Since(stime))

		}

	}

}
