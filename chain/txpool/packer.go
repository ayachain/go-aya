package txpool

import (
	"context"
	ASD "github.com/ayachain/go-aya/chain/sdaemon/common"
	AElectoral "github.com/ayachain/go-aya/vdb/electoral"
	"github.com/ayachain/go-aya/vdb/node"
	"github.com/ipfs/go-ipfs/pin"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/prometheus/common/log"
	"strings"
	"sync"
	"time"
)

func (pool *aTxPool) threadElectoralAndPacker ( ctx context.Context ) {

	log.Info("ATxPool Thread On: " + ATxPoolThreadTxPackage)
	defer log.Info("ATxPool Thread Off: " + ATxPoolThreadTxPackage)

	ctx1, cancel1 := context.WithCancel(ctx)
	ctx2, cancel2 := context.WithCancel(ctx)
	ctx3, cancel3 := context.WithCancel(ctx)

	awaiter := &sync.WaitGroup{}

	go subscribeThread( ctx1, pool, awaiter )
	go winerListnerThread( ctx2, pool, awaiter )
	go doPingsAndElectoral( ctx3, pool, awaiter )

	select {
	case <- ctx1.Done():
		break

	case <- ctx2.Done():
		break

	case <- ctx3.Done():
		break

	case <- ctx.Done():
		break
	}

	cancel1()
	cancel2()
	cancel3()
	awaiter.Wait()

	return
}

func subscribeThread( ctx context.Context, pool *aTxPool, awaiter *sync.WaitGroup ) {

	awaiter.Add(1)
	defer awaiter.Done()

	sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadTxPackage] )
	if err != nil {
		return
	}
	defer sub.Cancel()

	/// Recv other node vote
	for {

		msg, err := sub.Next(ctx)
		if err != nil {
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

}

func winerListnerThread( ctx context.Context, pool *aTxPool, awaiter *sync.WaitGroup ) {

	awaiter.Add(1)
	defer awaiter.Done()

	var lockPacker = true

	observerFunc := func(s ASD.Signal) {

		switch s {
		default:
			lockPacker = false

		case ASD.SignalOnConfirmed:

			if pool.lmblock != nil {

				pool.ind.Pinning.PinWithMode( pool.lmblock.Txs, pin.Any )

				txs := pool.lmblock.ReadTxsFromDAG(context.TODO(), pool.ind)

				_ = pool.confirmTxs( txs )
			}

			lockPacker = false
		}
	}

	/// chain core event observer
	_ = pool.asd.AddTimeoutObserver(observerFunc)
	defer pool.asd.RemoveObserver(observerFunc)

	for {

		packer, err := pool.eleservices.FightPacker(ctx)
		if err != nil {
			return
		}

		if packer == nil || lockPacker {
			continue
		}

		idx, err := pool.cvfs.Indexes().GetLatest()
		if err != nil {
			panic(err)
		}

		pool.asd.SendingSignal( idx.BlockIndex + 1, ASD.SignalDoPacking )

		if strings.EqualFold( packer.PackerPeerID, pool.ind.Identity.Pretty() ) {

			if pool.packerState == AElectoral.ATxPackStateLookup && packer.PackBlockIndex == idx.BlockIndex + 1 {

				if pool.GetState().Pending > 0 {

					pool.changePackerState(AElectoral.ATxPackStateMaster)

					if _, err := pool.doPackMBlock(); err != nil {

						log.Warn(err)
						continue

					}

					lockPacker = true
				}
			}

		} else {
			pool.changePackerState(AElectoral.ATxPackStateFollower)
		}
	}
}

func doPingsAndElectoral( ctx context.Context, pool *aTxPool, awaiter *sync.WaitGroup ) {

	awaiter.Add(1)
	defer awaiter.Done()

	pticker := time.NewTicker(time.Second * 5)
	defer pticker.Stop()

	//var timeOutPeers []*peer.ID
	//
	//var lepeer *peer.ID
	//
	//observerFunc := func(s ASD.Signal) {
	//
	//	switch s {
	//	case ASD.SignalDoPacking:
	//		timeOutPeers = append(timeOutPeers, lepeer)
	//
	//	case ASD.SignalOnConfirmed:
	//		timeOutPeers = make([]*peer.ID, 0)
	//	}
	//}
	//
	///// chain core event observer
	//_ = pool.asd.AddTimeoutObserver(observerFunc)
	//defer pool.asd.RemoveObserver(observerFunc)

	for {

		select {
		case <- pticker.C :

			superNodes := pool.cvfs.Nodes().GetSuperNodeList()

			wg := sync.WaitGroup{}

			for _, v := range superNodes {

				/// it's self
				if strings.EqualFold(v.PeerID, pool.ind.Identity.Pretty()) {
					pool.eleservices.UpdatePingRet( &node.PingRet{Node: v, RTT: 0, UTime: time.Now().Unix()} )
					continue
				}

				sctx, _ := context.WithCancel(ctx)
				wg.Add(1)

				go func(ctx context.Context, n *node.Node) {

					defer wg.Done()

					/// decode peer id
					pid, err := peer.IDB58Decode(n.PeerID)
					if err != nil {
						pool.eleservices.UpdatePingRet( &node.PingRet{Node: n, RTT: AElectoral.KPingTimeout, UTime: time.Now().Unix()} )
						return
					}

					if len( pool.ind.Peerstore.Addrs(pid) ) == 0 {

						sctx, cancel := context.WithTimeout(ctx, AElectoral.KPingTimeout)
						defer cancel()
						p, err := pool.ind.Routing.FindPeer(sctx, pid)

						if sctx.Err() != nil {
							return
						}

						if err != nil {
							pool.eleservices.UpdatePingRet( &node.PingRet{Node: n, RTT: AElectoral.KPingTimeout, UTime: time.Now().Unix()} )
							return
						} else {
							pool.ind.Peerstore.AddAddrs(p.ID, p.Addrs, peerstore.ConnectedAddrTTL)
						}
					}

					sctx, cancel := context.WithTimeout( ctx, AElectoral.KPingTimeout )
					defer cancel()
					r, ok := <- ping.Ping(sctx, pool.ind.PeerHost, pid)

					if sctx.Err() != nil {
						return
					}

					if !ok || r.Error != nil {
						pool.eleservices.UpdatePingRet( &node.PingRet{Node: n, RTT: AElectoral.KPingTimeout, UTime: time.Now().Unix()} )
						return
					}

					pool.eleservices.UpdatePingRet( &node.PingRet{Node: n, RTT: r.RTT, UTime: time.Now().Unix()} )

				}(sctx, v)
			}

			wg.Wait()

			if ctx.Err() != nil {
				return
			}

			if idx, err := pool.cvfs.Indexes().GetLatest(); err != nil {

				log.Warnf("DoElectoral:%v", err.Error())
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

				if err := pool.doBroadcast(vote, pool.channelTopics[ATxPoolThreadTxPackage]); err != nil {
					log.Warnf("DoElectoral:%v", err.Error())
				}
			}

		case <- ctx.Done():
			return
		}

	}
}