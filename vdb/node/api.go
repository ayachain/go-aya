package node

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
)

const DBPath = "/nodes"

type reader interface {

	GetNodeByPeerId( peerId string ) (*Node, error)

	GetSuperMaterTotalVotes() uint64

	GetSuperNodeList() []*Node
}

type writer interface {

	Insert( peerId string, node *Node ) error

	Del( peerId string )
}


type Services interface {

	AVdbComm.VDBSerices
	reader
}


type Caches interface {

	AVdbComm.VDBCacheServices

	reader
	writer
}