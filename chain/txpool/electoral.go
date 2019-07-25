package txpool

import (
	"context"
	"fmt"
	AElectoral "github.com/ayachain/go-aya/vdb/electoral"
	"github.com/ayachain/go-aya/vdb/node"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/pkg/errors"
	"go4.org/sort"
	"strings"
	"sync"
	"time"
)

var (
	PingExpected = time.Minute * 3600
)

func electoralThread(ctx context.Context) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadElectoral)

	pool := ctx.Value("Pool").(*ATxPool)

	subCtx, subCancel := context.WithCancel(ctx)

	pool.threadChans.Store(ATxPoolThreadElectoral, make(chan []byte, ATxPoolThreadElectoralBuff) )

	pool.workingThreadWG.Add(1)

	defer func() {

		subCancel()

		<- subCtx.Done()

		cc, exist := pool.threadChans.Load(ATxPoolThreadElectoral)

		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete(ATxPoolThreadElectoral)
		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadElectoral)

	}()


	go func() {

		sub, err := pool.ind.PubSub.Subscribe( pool.channelTopics[ATxPoolThreadElectoral] )

		if err != nil {
			return
		}

		for {

			msg, err := sub.Next(subCtx)

			if err != nil {
				return
			}

			/// msg from must be a super node
			if nd, err := pool.cvfs.Nodes().GetNodeByPeerId(msg.GetFrom().Pretty()); err != nil {

				continue

			} else {

				if nd.Type == node.NodeTypeSuper {

					cc, _ := pool.threadChans.Load(ATxPoolThreadElectoral)

					cc.(chan []byte) <- msg.Data
				}
			}
		}

	}()

	for {

		cc, _ := pool.threadChans.Load(ATxPoolThreadElectoral)

		select {
		case <- ctx.Done():

			return

		case rawmsg, isOpen := <- cc.(chan []byte):

			if !isOpen {
				continue
			}

			if pool.packerState != AElectoral.ATxPackStateLookup {
				continue
			}

			ele := &AElectoral.Electoral{}
			if err := ele.RawMessageDecode(rawmsg); err != nil {
				log.Error(err)
				continue
			}

			latestIndex, err := pool.cvfs.Indexes().GetLatest()
			if err != nil {
				log.Warning(err)
				continue
			}

			latestBlock, err := pool.cvfs.Blocks().GetBlocks(latestIndex.BlockIndex)
			if err != nil {
				log.Warning(err)
				continue
			}

			vmap, exist := pool.voteRetMapping.Load( ele.Address.String() )

			if !exist {

				if strings.EqualFold(ele.Address.String(), latestBlock[0].Packager ) {

					pool.voteRetMapping.Store(ele.Address.String(), map[string]int{ele.Address.String():80})

				} else {

					pool.voteRetMapping.Store(ele.Address.String(), map[string]int{ele.Address.String():100})
				}

				continue

			} else {

				var vm = vmap.(map[string]int)

				if strings.EqualFold(ele.Address.String(), latestBlock[0].Packager ) {

					vm[ele.Address.String()] = 80

				} else {

					vm[ele.Address.String()] = 100
				}

			}

			var votedSum = 0

			pool.voteRetMapping.Range(func(key, value interface{}) bool {

				vm, ok := value.(map[string]int)

				if !ok {
					return false
				}

				votedSum += len(vm)

				return true
			})

			if votedSum == len(pool.onlineSuperNode) {

				// electoral finished
				var master = ""
				var maxVoteCount = 0

				pool.voteRetMapping.Range(func(key, value interface{}) bool {

					vm, ok := value.(map[string]int)

					if !ok {
						return false
					}

					var total = 0

					for _, vmv := range vm {
						total += vmv
					}

					if total > maxVoteCount {
						maxVoteCount = total
						master = key.(string)
					}

					return true
				})

				if strings.EqualFold( master, pool.ind.Identity.Pretty() ) {

					pool.packerState = AElectoral.ATxPackStateNextMaster

				} else {

					pool.packerState = AElectoral.ATxPackStateFollower
				}

			}

		}
	}
}


type PingRet struct {
	Node *node.Node
	RTT time.Duration
}


func pingFastest( ctx context.Context ) <- chan *node.Node {

	replay := make(chan *node.Node, 1)

	pool := ctx.Value("Pool").(*ATxPool)

	superNodes := pool.cvfs.Nodes().GetSuperNodeList()

	if len( pool.ind.Peerstore.Addrs(pool.ind.Identity) ) == 0 {
		return nil
	}

	for _, v := range superNodes {

		if strings.EqualFold( v.PeerID, pool.ind.Identity.Pretty() ) {
			continue
		}

		go func( n *node.Node ) {

			pid, err := peer.IDB58Decode( n.PeerID )

			if err != nil {
				return
			}

			info, err := pool.ind.DHT.FindPeer( context.TODO(), pid )

			if err != nil {
				return
			}

			pool.ind.Peerstore.AddAddrs( info.ID, info.Addrs, pstore.ConnectedAddrTTL )

			ctx, cancel := context.WithTimeout( ctx, time.Second * 5 )
			defer cancel()

			r, ok := <- ping.Ping(ctx, pool.ind.PeerHost, info.ID)

			if !ok || r.Error != nil {
				return
			}

			replay <- v

		}( v )
	}

	return replay
}


func pings( ctx context.Context ) ([]PingRet, error) {

	pool := ctx.Value("Pool").(*ATxPool)

	superNodes := pool.cvfs.Nodes().GetSuperNodeList()

	if len( pool.ind.Peerstore.Addrs(pool.ind.Identity) ) == 0 {
		return nil, errors.New("self node not config address")
	}

	var pingrets []PingRet

	wg := sync.WaitGroup{}

	for _, v := range superNodes {

		if strings.EqualFold( v.PeerID, pool.ind.Identity.Pretty() ) {
			continue
		}

		wg.Add(1)
		go func( n *node.Node ) {

			pid, err := peer.IDB58Decode( n.PeerID )

			if err != nil {
				pingrets = append(pingrets, PingRet{Node:n, RTT:PingExpected})
				return
			}

			defer wg.Done()

			info, err := pool.ind.DHT.FindPeer( context.TODO(), pid )

			if err != nil {
				pingrets = append(pingrets, PingRet{Node:n, RTT:PingExpected})
				return
			}

			pool.ind.Peerstore.AddAddrs( info.ID, info.Addrs, pstore.ConnectedAddrTTL )

			ctx, cancel := context.WithTimeout( ctx, time.Second * 5 )
			defer cancel()

			r, ok := <- ping.Ping(ctx, pool.ind.PeerHost, info.ID)

			if !ok || r.Error != nil {
				pingrets = append(pingrets, PingRet{Node:n, RTT:PingExpected})
				return
			}

			pingrets = append(pingrets, PingRet{Node:n, RTT:r.RTT})

		}( v )

	}

	wg.Wait()

	sort.Slice( pingrets, func(i, j int) bool {
		return pingrets[i].RTT < pingrets[j].RTT
	})

	return pingrets, nil
}