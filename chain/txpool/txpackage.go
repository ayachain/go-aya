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
	"math/rand"
	"time"
)

const (
	PackageTxsLimit = 2048
)

var (
	PackageThreadSleepTime = time.Microsecond  * 100
)

func (pool *ATxPool) txPackageThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On : " + AtxThreadTxPackage)
	defer fmt.Println("ATxPool Thread Off : " + AtxThreadTxPackage)

	for {

		select {
		case <-ctx.Done():
			return

		default:

			if pool.miningBlock == nil && pool.enablePackTxThread {

				rand.Seed(time.Now().UnixNano())
				bindex, err := pool.cvfs.Indexes().GetLatest()
				if err != nil {
					pool.PowerOff(err)
					return
				}

				/// Because you need to wait for calculation, miningblock does not have a field for the final result.
				mblk := &AMsgMBlock.MBlock{}
				mblk.ExtraData = ""
				mblk.Index = bindex.BlockIndex + 1
				mblk.ChainID = pool.genBlock.ChainID
				mblk.Parent = bindex.Hash.String()
				mblk.Timestamp = uint64(time.Now().Unix())
				mblk.RandSeed = rand.Int31()

				count := uint16(0)
				var txs []ATx.Transaction

				sshot, err := pool.storage.GetSnapshot()
				if err != nil {
					pool.PowerOff(err)
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
					pool.PowerOff(err)
					return
				}

				iblk := blocks.NewBlock(txsblockcontent)
				err = pool.ind.Blocks.AddBlock(iblk)
				if err != nil {
					pool.PowerOff(err)
					return
				}

				//packing
				mblk.Txc = count
				mblk.Txs = iblk.Cid().String()

				if err := pool.storage.Write(batch, nil); err != nil {
					break
				}

				pool.miningBlock = mblk

				if err := pool.DoBroadcast(mblk); err != nil {
					break
				}

				pool.enablePackTxThread = false

			} else {

				time.Sleep(PackageThreadSleepTime)

			}

		}

	}

}