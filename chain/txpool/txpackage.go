package txpool

import (
	"context"
	"encoding/json"
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	blocks "github.com/ipfs/go-block-format"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/rand"
	"time"
)

const (
	PackageTxsLimit = 1024
)

var (
	PackageThreadSleepTime = time.Microsecond  * 100
)

func (pool *ATxPool) txPackageThread(ctx context.Context) {

	fmt.Println("ATxPool txpackage thread power on")
	defer fmt.Println("ATxPool txpackage thread power off")

	for {

		select {
		case <-ctx.Done():
			return

		default:

			if pool.miningBlock == nil && !pool.IsEmpty() {

				poolTransaction, err := pool.storage.OpenTransaction()

				if err != nil {
					pool.PowerOff(err)
					return
				}

				rand.Seed(time.Now().UnixNano())
				bestBlock := pool.cvfs.Blocks().BestBlock()

				/// Because you need to wait for calculation, miningblock does not have a field for the final result.
				mblk := &AMsgMBlock.MBlock{}
				mblk.ExtraData = ""
				mblk.Index = bestBlock.Index + 1
				mblk.ChainID = bestBlock.ChainID
				mblk.Parent = bestBlock.GetHash().Hex()
				mblk.Timestamp = uint64(time.Now().Unix())
				mblk.RandSeed = rand.Int31()

				count := uint16(0)
				var txs []ATx.Transaction
				it := poolTransaction.NewIterator(&util.Range{Start: TxHashIteratorStart, Limit: TxHashIteratorLimit}, nil)

				for it.Next() {

					signedmsg, err := AKeyStore.BytesToRawMsg(it.Value())
					if err != nil {
						poolTransaction.Delete(it.Key(), nil)
						continue
					}

					if signedmsg.Content[0] != ATx.MessagePrefix {
						poolTransaction.Delete(it.Key(), nil)
						continue
					}

					subTx := ATx.Transaction{}
					if err = subTx.Decode(signedmsg.Content[1:]); err != nil {
						poolTransaction.Delete(it.Key(), nil)
						continue
					}

					txs = append(txs, subTx)
					count++

					err = poolTransaction.Delete(it.Key(), nil)
					if err != nil {
						pool.PowerOff(err)
					}

					if count >= PackageTxsLimit {
						break
					}
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

				if err := poolTransaction.Commit(); err != nil {
					break
				}

				pool.miningBlock = mblk

				fmt.Println( "PackageMiningBlockHash:" + pool.miningBlock.GetHash().String() )

				if err := pool.DoBroadcast(mblk); err != nil {
					break
				}

				c, exist := pool.threadChans[AtxThreadsNameMining]
				if exist {

					signmsg, err := pool.sign(mblk)
					if err != nil {
						break
					}

					c <- signmsg
				}

			} else {
				time.Sleep(PackageThreadSleepTime)
			}

		}

	}

}