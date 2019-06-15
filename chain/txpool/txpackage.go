package txpool

import (
	"context"
	"encoding/json"
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ayachain/go-aya/vdb/common"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	blocks "github.com/ipfs/go-block-format"
	"github.com/syndtr/goleveldb/leveldb"
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

	fmt.Println("ATxPool Thread On : " + AtxThreadTxPackage)
	defer fmt.Println("ATxPool Thread Off : " + AtxThreadTxPackage)

	for {

		select {
		case <-ctx.Done():
			return

		default:

			if pool.miningBlock == nil && !pool.IsEmpty() {

				pool.txLocker.Lock()

				rand.Seed(time.Now().UnixNano())
				bindex := pool.cvfs.Indexes().GetLatest()

				/// Because you need to wait for calculation, miningblock does not have a field for the final result.
				mblk := &AMsgMBlock.MBlock{}
				mblk.ExtraData = ""
				mblk.Index = bindex.BlockIndex + 1
				mblk.ChainID = pool.chainId
				mblk.Parent = bindex.Hash.String()
				mblk.Timestamp = uint64(time.Now().Unix())
				mblk.RandSeed = rand.Int31()

				count := uint16(0)
				var txs []ATx.Transaction

				it := pool.storage.NewIterator(&util.Range{Start: TxHashIteratorStart, Limit: TxHashIteratorLimit}, nil)

				loopCount := uint64(0)

				batch := &leveldb.Batch{}

				for it.Next() {

					loopCount ++
					signedmsg, err := AKeyStore.BytesToRawMsg(it.Value())
					if err != nil {
						batch.Delete(it.Key())
						continue
					}

					if signedmsg.Content[0] != ATx.MessagePrefix {
						batch.Delete(it.Key())
						continue
					}

					subTx := ATx.Transaction{}
					if err = subTx.Decode(signedmsg.Content[1:]); err != nil {
						batch.Delete(it.Key())
						continue
					}

					txcount, err := pool.cvfs.Transactions().GetTxCount(subTx.From)
					if err != nil {
						it.Release()
						panic(err)
					}

					if subTx.Tid - txcount != 1 {
						// need waiting join then queue
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

				newSize := pool.Size() - loopCount
				if newSize <= 0 {
					batch.Put(TxCount, common.BigEndianBytes( uint64(0) ))

				} else {
					batch.Put(TxCount, common.BigEndianBytes( newSize ))
				}

				//commit block to ipfs block
				txsblockcontent, err := json.Marshal(txs)
				if err != nil {
					pool.PowerOff(err)
					pool.txLocker.Unlock()
					return
				}

				iblk := blocks.NewBlock(txsblockcontent)
				err = pool.ind.Blocks.AddBlock(iblk)
				if err != nil {
					pool.PowerOff(err)
					pool.txLocker.Unlock()
					return
				}

				//packing
				mblk.Txc = count
				mblk.Txs = iblk.Cid().String()

				if err := pool.storage.Write(batch, nil); err != nil {
					pool.txLocker.Unlock()
					break
				}

				pool.miningBlock = mblk

				if err := pool.DoBroadcast(mblk); err != nil {
					pool.txLocker.Unlock()
					break
				}

				pool.txLocker.Unlock()

			} else {

				time.Sleep(PackageThreadSleepTime)

			}

		}

	}

}