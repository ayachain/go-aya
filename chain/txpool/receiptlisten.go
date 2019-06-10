package txpool

import (
	"context"
	"encoding/json"
	"fmt"
	AMsgMinied "github.com/ayachain/go-aya/vdb/minined"
	"github.com/syndtr/goleveldb/leveldb"
)

func (pool *ATxPool) receiptListen(ctx context.Context) {

	fmt.Println("ATxPool minined receipt thread power on")
	defer fmt.Println("ATxPool minined receipt thread power off")

	for {

		select {

		case rmsg := <- pool.threadChans[AtxThreadsNameReceiptListen] :

			pool.packLocker.Lock()
			defer pool.packLocker.Unlock()

			if rmsg.Content[0] != AMsgMinied.MessagePrefix {
				break
			}

			from, err := rmsg.ECRecover()
			fmt.Println("RCPAddr:" + from.String())
			if err != nil {
				break
			}

			fromVoting, err := pool.cvfs.Assetses().VotingCountOf(from.Bytes())
			if err != nil {
				break
			}

			rcp := &AMsgMinied.Minined{}
			err = rcp.RawMessageDecode( rmsg.Content )
			if err != nil {
				break
			}

			fmt.Println( "TxPoolMiningBlockHash:" + pool.miningBlock.GetHash().String() )
			fmt.Println( "RcpMBlockHash:" + rcp.MBlockHash.String() )

			if pool.miningBlock.GetHash() != rcp.MBlockHash {
				break
			}

			receiptKey := []byte(TxReceiptPrefix)
			copy( receiptKey[1:], rcp.MBlockHash.Bytes() )

			exist, err := pool.storage.Has(receiptKey, nil)
			if err != nil {
				pool.Close()
				pool.KillByErr(ErrStorageLowAPIExpected)
			}

			rcidstr := rcp.RetCID.String()
			receiptMap := make(map[string]uint64)
			if exist {

				value, err := pool.storage.Get(receiptKey, nil)
				if err != nil {
					pool.KillByErr(ErrStorageLowAPIExpected)
				}

				if json.Unmarshal(value, receiptMap) != nil {
					pool.KillByErr(ErrStorageRawDataDecodeExpected)
				}

				ocount, vexist := receiptMap[rcidstr]
				if vexist {
					receiptMap[rcidstr] = ocount + fromVoting
				} else {
					receiptMap[rcidstr] = fromVoting
				}

			} else {
				receiptMap[rcidstr] = fromVoting
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

				pool.miningBlockLocker.Lock()
				defer pool.packLocker.Unlock()

				batch := &leveldb.Batch{}
				batch.Delete(receiptKey)
				batch.Delete(rcp.MBlockHash.Bytes())

				if err := pool.storage.Write(batch, nil); err != nil {
					pool.PowerOff(err)
					return
				}

				pool.miningBlock.ExtraData = rcidstr

			}

		}

	}

}
