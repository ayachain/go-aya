package txpool

import (
	"context"
	"encoding/json"
	"fmt"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	blocks "github.com/ipfs/go-block-format"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"time"
)

func txPackageThread(ctx context.Context ) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadTxPackage)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.tcmapMutex.Lock()
	pool.threadChans[ATxPoolThreadTxPackage] = make(chan []byte, ATxPoolThreadTxPackageBuff)
	pool.tcmapMutex.Unlock()

	defer func() {

		pool.tcmapMutex.Lock()
		cc, exist := pool.threadChans[ATxPoolThreadTxPackage]

		if exist {

			close( cc )
			delete(pool.threadChans, ATxPoolThreadTxPackage)

		}
		pool.tcmapMutex.Unlock()

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadTxPackage)

	}()


	for {

		select {
		case <- ctx.Done():
			return

		case _, isOpen := <- pool.threadChans[ATxPoolThreadTxPackage]:

			if !isOpen {
				continue
			}

			if pool.miningBlock != nil {
				continue
			}

			bindex, err := pool.cvfs.Indexes().GetLatest()
			if err != nil {
				log.Error(err)
				return
			}

			/// Because you need to wait for calculation, miningblock does not have a field for the final result.
			mblk := &AMsgMBlock.MBlock{}
			mblk.ExtraData = ""
			mblk.Index = bindex.BlockIndex + 1
			mblk.ChainID = pool.genBlock.ChainID
			mblk.Parent = bindex.Hash.String()
			mblk.Timestamp = uint64(time.Now().Unix())

			count := uint16(0)
			var txs []ATx.Transaction

			sshot, err := pool.storage.GetSnapshot()
			if err != nil {
				log.Error(err)
				return
			}

			it := sshot.NewIterator(&util.Range{Start: TxHashIteratorStart, Limit: TxHashIteratorLimit}, nil)

			batch := &leveldb.Batch{}

			for it.Next() {

				subTx := ATx.Transaction{}

				if err := subTx.Decode(it.Value()); err != nil {
					batch.Delete(it.Key())
					continue
				}

				txs = append(txs, subTx)
				count++

				batch.Delete(it.Key())

				if count >= PackageTxsLimit {
					break
				}

			}

			it.Release()

			if count <= 0 {
				continue
			}

			//commit block to ipfs block
			txsblockcontent, err := json.Marshal(txs)
			if err != nil {
				log.Error(err)
				return
			}

			iblk := blocks.NewBlock(txsblockcontent)
			err = pool.ind.Blocks.AddBlock(iblk)
			if err != nil {
				log.Error(err)
				return
			}

			//packing
			mblk.Txc = count
			mblk.Txs = iblk.Cid().String()

			if err := pool.storage.Write(batch, nil); err != nil {
				log.Error(err)
				return
			}

			pool.miningBlock = mblk

			if err := pool.doBroadcast(mblk, pool.channelTopics[ATxPoolThreadMining]); err != nil {
				log.Error(err)
				return
			}
		}

	}

}