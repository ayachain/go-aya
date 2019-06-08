package txpool

import (
	"context"
	"encoding/json"
	AMsgMBlock "github.com/ayachain/go-aya/chain/message/miningblock"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	blocks "github.com/ipfs/go-block-format"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/rand"
	"time"
)

/// Super Node Mode Specialization
/// The super node is responsible for the package and broadcast of the transaction
/// after receiving the transaction.
/// The block data sent must be an undetermined block, which needs to be processed
/// by other nodes. We define the message header as "m", which is actually the meaning
/// of minerblock. We tell other nodes that I want to use this piece to help me calculate
/// the final result.
const PackageTxsLimit = 1024

var (
	WarningMiningHashWaitingReceipt = errors.New("there are data that are being calculated and cannot be submitted repeatedly.")
)

func (pool *ATxPool) threadTransactionPackage(pctx context.Context) <- chan error {

	replay := make(chan error)

	go func() {

		for {

			select {
			case <- pctx.Done() :
				break

			case <- pool.doPackingMiningBlock :

				/// If there is an undetermined block that is being computed when a package request
				/// is initiated, it needs to wait. In fact, this is wrong. This is not allowed in
				/// the logic control of txpool. That is to say, if there is a block currently being
				/// computed, it is allowed to pack the next block.
				for pool.miningBlock != nil {
					replay <- WarningMiningHashWaitingReceipt
				}

				poolTransaction, err := pool.storage.OpenTransaction()
				if err != nil {
					replay <- err
					break
				}

				rand.Seed(time.Now().UnixNano())

				mblk := &AMsgMBlock.MsgRawMiningBlock{}

				/// Because you need to wait for calculation, miningblock does not have a field for the final result.
				mblk.ExtraData = ""

				mblk.Index = pool.bestBlock.Index
				mblk.ChainID = pool.chainId
				mblk.Parent = pool.bestBlock.GetHash().Hex()
				mblk.Timestamp = uint64(time.Now().Unix())
				mblk.RandSeed = rand.Int31()


				it := poolTransaction.NewIterator(&util.Range{Start:TxHashIteratorStart,Limit:TxHashIteratorLimit}, nil)

				count := uint16(0)

				var txs []ATx.Transaction
				var looperr error

				for it.Next(){

					subTx := ATx.Transaction{}

					if looperr = subTx.Decode( it.Value() ); err != nil {
						goto loopend
					}

					txs = append(txs, subTx)

					count ++

					err := poolTransaction.Delete(it.Key(), nil)
					if err != nil {
						pool.KillByErr(err)
					}

					if count >= PackageTxsLimit {
						break
					}
				}

				loopend:

					if looperr != nil {

						replay <- looperr
						break

					} else {

						//commit block to ipfs block
						txsblockcontent, err := json.Marshal(txs)
						if err != nil {
							replay <- err
							break
						}

						iblk := blocks.NewBlock( txsblockcontent )
						err = pool.ind.Blocks.AddBlock(iblk)
						if err != nil {
							replay <- err
							break
						}

						//packing
						mblk.Txc = count
						mblk.Txs = iblk.Cid().String()

						if err := poolTransaction.Commit(); err != nil {
							replay <- err
							break
						}

						pool.miningBlock = mblk
						replay <- nil
						break
					}
			}

		}

	}()

	return replay
}