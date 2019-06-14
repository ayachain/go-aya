package txpool

import (
	"context"
	"encoding/json"
	"fmt"
	AMsgMinied "github.com/ayachain/go-aya/vdb/minined"
	"github.com/syndtr/goleveldb/leveldb"
)

func (pool *ATxPool) receiptListen(ctx context.Context) {

	fmt.Println("ATxPool Thread On : " + AtxThreadReceiptListen)
	defer fmt.Println("ATxPool Thread Off : " + AtxThreadReceiptListen)

	for {

		select {

		case rmsg := <- pool.threadChans[AtxThreadReceiptListen] :

			//pool.packLocker.Lock()
			//defer pool.packLocker.Unlock()

			if rmsg.Content[0] != AMsgMinied.MessagePrefix {
				break
			}

			from, err := rmsg.ECRecover()
			if err != nil {
				break
			}

			fromAst, err := pool.cvfs.Assetses().AssetsOf(*from)
			if err != nil {
				break
			}

			rcp := &AMsgMinied.Minined{}
			err = rcp.RawMessageDecode( rmsg.Content )
			if err != nil {
				break
			}

			if pool.miningBlock.GetHash() != rcp.MBlockHash {
				break
			}

			receiptKey := []byte(TxReceiptPrefix)
			copy( receiptKey[1:], rcp.MBlockHash.Bytes() )

			exist, err := pool.storage.Has(receiptKey, nil)
			if err != nil {
				pool.Close()
				pool.PowerOff(ErrStorageLowAPIExpected)
			}

			rcidstr := rcp.RetCID.String()
			receiptMap := make(map[string]uint64)
			if exist {

				value, err := pool.storage.Get(receiptKey, nil)
				if err != nil {
					pool.PowerOff(ErrStorageLowAPIExpected)
				}

				if json.Unmarshal(value, receiptMap) != nil {
					pool.PowerOff(ErrStorageRawDataDecodeExpected)
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
				pool.PowerOff(ErrStorageLowAPIExpected)
				return
			}

			if err := pool.storage.Put( receiptKey, rmapbs, nil ); err != nil {
				pool.PowerOff(ErrStorageLowAPIExpected)
				return
			}


			if receiptMap[rcidstr] > pool.ownerAsset.Vote * 3 || pool.workmode == AtxPoolWorkModeOblivioned {

				batch := &leveldb.Batch{}
				batch.Delete(receiptKey)
				batch.Delete(rcp.MBlockHash.Bytes())

				if err := pool.storage.Write(batch, nil); err != nil {
					pool.PowerOff(err)
					return
				}


				cblock := pool.miningBlock.Confirm(rcidstr)

				if err := pool.DoBroadcast(cblock); err != nil {
					break
				}

				//c, exist := pool.threadChans[AtxThreadsNameBlockPacage]
				//if exist {
				//
				//	signmsg, err := pool.sign(cblock)
				//
				//	if err != nil {
				//		break
				//	}
				//
				//	c <- signmsg
				//}
			}

		}

	}

}
