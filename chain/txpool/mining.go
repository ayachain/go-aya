package txpool

import (
	"context"
	"fmt"
	"github.com/ayachain/go-aya/consensus/core"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMsgMined "github.com/ayachain/go-aya/vdb/minined"
	"github.com/ayachain/go-aya/vdb/node"
	IBlocks "github.com/ipfs/go-block-format"
	"strings"
)

func miningThread(ctx context.Context ) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadMining)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans.Store(ATxPoolThreadMining, make(chan []byte, ATxPoolThreadTxPackageBuff) )

	subCtx, subCancel := context.WithCancel(ctx)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans.Load(ATxPoolThreadMining)
		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete(ATxPoolThreadMining)
		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadMining)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadMining] )

		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				return
			}

			if pool.workmode == AtxPoolWorkModeSuper {

				nd, err := pool.cvfs.Nodes().GetNodeByPeerId( msg.GetFrom().Pretty() )

				if err != nil || nd.Type != node.NodeTypeSuper {
					/// TODO dissconnect this node
					continue
				}

				mblock := &AMsgMBlock.MBlock{}

				if err := mblock.RawMessageDecode(msg.Data); err != nil {
					log.Error(err)
					continue
				}

				if pool.miningBlock != nil && !strings.EqualFold(pool.miningBlock.GetHash().String(), mblock.GetHash().String()) {

					continue

				} else {

					if pool.miningBlock == nil {

						pool.miningBlock = mblock

					} else if pool.miningBlock.Index < mblock.Index {

						pool.miningBlock = mblock

					} else {

						continue

					}

				}

				if strings.EqualFold( msg.GetFrom().Pretty(), pool.eleservices.LatestPacker().PackerPeerID ) &&
					!strings.EqualFold( msg.GetFrom().Pretty(), pool.ind.Identity.Pretty() ) {

					if err := pool.doBroadcast(mblock, pool.channelTopics[ATxPoolThreadMining] ); err != nil {
						continue
					}

				}

			} else {

				if <- pool.notary.TrustOrNot(msg, core.NotaryMessageMiningBlock, pool.cvfs) {

					cc, _ := pool.threadChans.Load(ATxPoolThreadMining)

					cc.(chan []byte) <- msg.Data

				}
			}

		}

	}()


	for {

		cc, _ := pool.threadChans.Load(ATxPoolThreadMining)

		select {

		case <- ctx.Done():
			return

		case rawmsg, isOpen := <- cc.(chan []byte):

			if !isOpen {
				continue
			}

			mblock := &AMsgMBlock.MBlock{}
			if err := mblock.RawMessageDecode(rawmsg); err != nil {
				log.Error(err)
				continue
			}

			cVFS, err := pool.cvfs.NewCVFSCache()
			if err != nil {
				log.Error(err)
				return
			}

			group, err := pool.notary.MiningBlock(mblock, cVFS)
			if err != nil {

				if err := cVFS.Close(); err != nil {
					log.Error(err)
					return
				}

				continue
			}

			if err := cVFS.Close(); err != nil {
				log.Error(err)
				return
			}

			groupbytes := group.Encode()
			gblock := IBlocks.NewBlock(groupbytes)
			if err := pool.ind.Blocks.AddBlock(gblock); err != nil {
				log.Error(err)
				return
			}

			mRet := &AMsgMined.Minined{
				MBlockHash:mblock.GetHash(),
				RetCID:gblock.Cid(),
			}

			if err := pool.doBroadcast(mRet, pool.channelTopics[ATxPoolThreadReceiptListen]); err != nil {
				log.Error(err)
				return
			}
		}
	}
}
