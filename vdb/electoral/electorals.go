package electoral

import (
	"context"
	"github.com/ayachain/go-aya/vdb"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/node"
	EComm "github.com/ethereum/go-ethereum/common"
	"sync"
	"time"
)


type aElectorals struct {

	MemServices

	vfs vdb.CVFS

	exp int64

	votoMapping map[string]*im.Electoral
	votoMu sync.Mutex

	pingMapping map[string]*node.PingRet
	pingMu sync.Mutex

	currnetMaster string

	packerChan chan *EleRet

	latestPacker *EleRet
}


func CreateServices( cvfs vdb.CVFS, exptime int64 ) MemServices {

	return &aElectorals{
		vfs:cvfs,
		exp:exptime,
		packerChan:make(chan *EleRet),
		votoMapping:make(map[string]*im.Electoral),
		pingMapping:make(map[string]*node.PingRet),
	}

}

func (aele *aElectorals) UpdateVote( electoral *im.Electoral ) {

	aele.votoMu.Lock()
	defer aele.votoMu.Unlock()

	aele.votoMapping[ EComm.BytesToAddress(electoral.From).String() ] = electoral

	//log.Infof( "From:%v -> PeerID:%v", electoral.From.String(), electoral.ToPeerId )

	onlineNodeCount := 0
	superNodesCount := int(aele.vfs.Nodes().GetSuperNodeCount())

	for k := range aele.pingMapping {

		spret := aele.pingMapping[k]

		if time.Now().Unix() - spret.UTime <= 10 && spret.RTT != KPingTimeout {
			onlineNodeCount++
		}

	}

	passCount := superNodesCount / 2 + 1

	if onlineNodeCount < passCount {
		return
	}

	lidx, err := aele.vfs.Indexes().GetLatest()
	if err != nil {
		return
	}

	vmap := make(map[string]int)

	for _, e := range aele.votoMapping {

		if e.BlockIndex == lidx.BlockIndex + 1 {
			vmap[e.ToPeerId] ++
		}

	}

	var winner string
	var vcount = 0

	for k, v := range vmap {

		if v > vcount {
			winner = k
			vcount = v
		}

	}

	if vcount >= passCount {

		aele.latestPacker = &EleRet{ lidx.BlockIndex + 1, winner }

		aele.packerChan <- aele.latestPacker
	}

}


func (aele *aElectorals) UpdatePingRet( pret *node.PingRet ) {

	aele.pingMu.Lock()
	defer aele.pingMu.Unlock()

	aele.pingMapping[pret.Node.PeerID] = pret
}


func (aele *aElectorals) GetNearestOnlineNode( bindex uint64 ) *im.Node {

	nds := aele.vfs.Nodes().GetSuperNodeList()

	idx := bindex % uint64(len(nds))

	for i := idx; i < uint64(len(nds)); i++ {

		tnd := nds[i]

		if ping, exist := aele.pingMapping[tnd.PeerID]; !exist {

			continue

		} else {

			if ping.RTT != KPingTimeout {
				return tnd
			}

		}

	}

	for i := uint64(0); i < uint64(len(nds)); i++ {

		tnd := nds[i]

		if ping, exist := aele.pingMapping[tnd.PeerID]; !exist {

			continue

		} else {

			if ping.RTT != KPingTimeout {
				return tnd
			}

		}

	}

	return nil
}

func (aele *aElectorals) GetOnlineSuperNodes() []string {

	aele.pingMu.Lock()
	defer aele.pingMu.Unlock()

	var ret []string

	for k := range aele.pingMapping {

		spret := aele.pingMapping[k]

		if time.Now().Unix() - spret.UTime <= 10 && spret.RTT != KPingTimeout {
			ret = append(ret, k)
		}

	}

	return ret
}

func (aele *aElectorals) GetNodesPingStates() []*ConnState {

	aele.pingMu.Lock()
	defer aele.pingMu.Unlock()

	snds := aele.vfs.Nodes().GetSuperNodeList()

	var rets []*ConnState

	for _, nd := range snds {

		if pret, exist := aele.pingMapping[nd.PeerID]; exist {

			if ele, ok := aele.votoMapping[nd.PeerID]; !ok {

				rets = append(rets, &ConnState {
					BestBlockIndex:0,
					OwnerAddress:EComm.BytesToAddress(nd.Owner),
					PID:nd.PeerID,
					RTT:pret.RTT,
				})

			} else {

				rets = append(rets, &ConnState{

					BestBlockIndex:ele.BestIndex,
					OwnerAddress:EComm.BytesToAddress(nd.Owner),
					PID:nd.PeerID,
					RTT:pret.RTT,
				})
			}

		} else {

			rets = append(rets, &ConnState {

				BestBlockIndex:0,
				OwnerAddress:EComm.BytesToAddress(nd.Owner),
				PID:nd.PeerID,
				RTT:KPingTimeout,
			})

		}

	}

	return rets

}

func (aele *aElectorals) FightPacker(ctx context.Context) (*EleRet, error) {

	select {
	case <- ctx.Done() :
		return nil, ctx.Err()

	case r := <- aele.packerChan :
		return r, nil
	}

}


func (aele *aElectorals) LatestPacker() *EleRet {
	return aele.latestPacker
}