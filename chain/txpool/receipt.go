package txpool

import (
	"context"
	"encoding/json"
	"fmt"
	AMsgMinied "github.com/ayachain/go-aya/vdb/minined"
	"github.com/syndtr/goleveldb/leveldb"
)

func (pool *ATxPool) receiptListen(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + AtxThreadReceiptListen)

	tctx, cancel := context.WithCancel(ctx)

	pool.threadClosewg.Add(1)

	for {

		select {
		case <- tctx.Done():

			fmt.Println("ATxPool Thread Off: " + AtxThreadReceiptListen)

			pool.threadClosewg.Done()

			return

		case rmsg, isOpen := <- pool.threadChans[AtxThreadReceiptListen] :

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
				cancel()
				continue
			}

			rcidstr := rcp.RetCID.String()
			receiptMap := make(map[string]uint64)
			if exist {

				value, err := pool.storage.Get(receiptKey, nil)
				if err != nil {
					log.Error(ErrStorageLowAPIExpected)
					cancel()
					continue
				}

				if json.Unmarshal(value, receiptMap) != nil {
					log.Error(ErrStorageRawDataDecodeExpected)
					cancel()
					continue
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
				cancel()
				continue
			}

			if err := pool.storage.Put( receiptKey, rmapbs, nil ); err != nil {
				log.Error(ErrStorageLowAPIExpected)
				cancel()
			}


			if receiptMap[rcidstr] > pool.ownerAsset.Vote * 3 || pool.workmode == AtxPoolWorkModeOblivioned {

				batch := &leveldb.Batch{}
				batch.Delete(receiptKey)
				batch.Delete(rcp.MBlockHash.Bytes())

				if err := pool.storage.Write(batch, nil); err != nil {
					log.Error(err)
					cancel()
					continue
				}


				cblock := pool.miningBlock.Confirm(rcidstr)

				if err := pool.DoBroadcast(cblock); err != nil {
					log.Error(err)
					cancel()
					continue
				}

			}

		}

	}

}
