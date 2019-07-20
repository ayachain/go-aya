package txpool

import (
	"context"
	"encoding/json"
	"fmt"
	AElectoral "github.com/ayachain/go-aya/vdb/electoral"
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

	pool.threadChans.Store(ATxPoolThreadTxPackage, make(chan []byte, ATxPoolThreadTxPackageBuff))

	defer func() {

		cc, exist := pool.threadChans.Load(ATxPoolThreadTxPackage)
		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete(ATxPoolThreadTxPackage)

		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadTxPackage)

	}()


	for {

		cc, _ := pool.threadChans.Load(ATxPoolThreadTxPackage)

		select {
		case <- ctx.Done():
			return

		case _, isOpen := <- cc.(chan []byte):

			if !isOpen {
				continue
			}

			if pool.miningBlock != nil {
				continue
			}

			/// If the current node is voted "Master", do packing mining block.
			if pool.packerState != AElectoral.ATxPackStateMaster {
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
			mblk.Packager = pool.ownerAccount.Address.String()

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