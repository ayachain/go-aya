package txpool

import (
	"context"
	"fmt"
	AElectoral "github.com/ayachain/go-aya/vdb/electoral"
	"github.com/ayachain/go-aya/vdb/node"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"strings"
	"sync"
	"time"
)

func electoralThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadElectoral)

	pool := ctx.Value("Pool").(*ATxPool)

	subCtx, subCancel := context.WithCancel(ctx)

	pctx := context.WithValue(ctx, "Pool", pool)

	fctx, fcancel := context.WithCancel(ctx)

	pool.threadChans.Store(ATxPoolThreadElectoral, make(chan []byte, ATxPoolThreadElectoralBuff) )

	pool.workingThreadWG.Add(1)

	defer func() {

		subCancel()

		fcancel()

		<- fctx.Done()

		<- pctx.Done()

		<- subCtx.Done()

		cc, exist := pool.threadChans.Load(ATxPoolThreadElectoral)

		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete(ATxPoolThreadElectoral)
		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadElectoral)

	}()


	/// Recv other node vote
	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadElectoral] )

		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				log.Warning(err)
				return
			}

			/// msg from must be a super node
			if nd, err := pool.cvfs.Nodes().GetNodeByPeerId(msg.GetFrom().Pretty()); err != nil {

				continue

			} else {

				if nd.Type == node.NodeTypeSuper {

					ele := &AElectoral.Electoral{}
					if err := ele.RawMessageDecode(msg.Data); err != nil {
						log.Error(err)
						continue
					}

					pool.eleservices.UpdateVote( ele )
				}
			}
		}

	}()

	go func() {

		for {

			select {
			case <- fctx.Done():

				return

			case packer := <- pool.eleservices.FightPacker():

				if packer == nil {
					continue
				}

				if strings.EqualFold( packer.PackerPeerID, pool.ind.Identity.Pretty() ) {

					if idx, err := pool.cvfs.Indexes().GetLatest(); err != nil {

						continue

					} else {

						if pool.packerState == AElectoral.ATxPackStateLookup && packer.PackBlockIndex == idx.BlockIndex + 1 {

							if pool.txlen > 0 {
								pool.changePackerState(AElectoral.ATxPackStateMaster)
								pool.DoPackMBlock()
							}

						}
					}

				} else {

					pool.changePackerState(AElectoral.ATxPackStateFollower)

				}
			}

		}

	}()

	doPingsAndElectoral(pctx)
}


func doPingsAndElectoral( ctx context.Context ) {

	pool := ctx.Value("Pool").(*ATxPool)

	for {

		select {

		case <- ctx.Done():

			return

		default:

			time.Sleep( time.Second * 5 )

		}

		superNodes := pool.cvfs.Nodes().GetSuperNodeList()

		wg := sync.WaitGroup{}

		for _, v := range superNodes {

			if strings.EqualFold(v.PeerID, pool.ind.Identity.Pretty()) {
				continue
			}

			wg.Add(1)

			go func(n *node.Node) {

				defer wg.Done()

				pid, err := peer.IDB58Decode(n.PeerID)

				if err != nil {
					pool.eleservices.UpdatePingRet( &node.PingRet{Node: n, RTT: AElectoral.KPingTimeout, UTime: time.Now().Unix()} )
					return
				}

				if len( pool.ind.Peerstore.Addrs(pid) ) == 0 {

					ctx, cancel := context.WithTimeout(context.TODO(), AElectoral.KPingTimeout)

					p, err := pool.ind.Routing.FindPeer(ctx, pid)

					cancel()

					if err != nil {

						pool.eleservices.UpdatePingRet( &node.PingRet{Node: n, RTT: AElectoral.KPingTimeout, UTime: time.Now().Unix()} )

						return

					} else {

						pool.ind.Peerstore.AddAddrs(p.ID, p.Addrs, pstore.ConnectedAddrTTL)

					}
				}

				ctx, cancel := context.WithTimeout( ctx, AElectoral.KPingTimeout )

				r, ok := <- ping.Ping(ctx, pool.ind.PeerHost, pid)

				cancel()

				if !ok || r.Error != nil {
					pool.eleservices.UpdatePingRet( &node.PingRet{Node: n, RTT: AElectoral.KPingTimeout, UTime: time.Now().Unix()} )
					return
				}

				pool.eleservices.UpdatePingRet( &node.PingRet{Node: n, RTT: r.RTT, UTime: time.Now().Unix()} )

			}(v)
		}

		wg.Wait()

		if idx, err := pool.cvfs.Indexes().GetLatest(); err != nil {
			log.Warningf("DoElectoral:%v", err.Error())
			continue

		} else {

			packer := pool.eleservices.GetNearestOnlineNode( idx.BlockIndex + 1 )

			if packer == nil {
				continue
			}

			vote := &AElectoral.Electoral {
				BestIndex:idx.BlockIndex,
				BlockIndex:idx.BlockIndex + 1,
				From:pool.ownerAccount.Address,
				ToPeerId: packer.PeerID,
				Time:time.Now().Unix(),
			}

			if err := pool.doBroadcast(vote, pool.channelTopics[ATxPoolThreadElectoral]); err != nil {
				log.Warningf("DoElectoral:%v", err.Error())
				continue
			}

		}

	}
}