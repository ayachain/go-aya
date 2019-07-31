package electoral

import (
	"github.com/ayachain/go-aya/vdb/node"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

const KPingTimeout = 10 * time.Second

type EleRet struct {
	PackBlockIndex uint64
	PackerPeerID string
}

type ConnState struct {
	BestBlockIndex uint64
	OwnerAddress common.Address
	PID string
	RTT time.Duration
}

type MemServices interface {

	UpdateVote( electoral *Electoral )

	UpdatePingRet( pret *node.PingRet )

	GetOnlineSuperNodes() []string

	GetNodesPingStates() []*ConnState

	GetNearestOnlineNode( bindex uint64 ) *node.Node

	FightPacker() <- chan *EleRet

	LatestPacker() *EleRet
}