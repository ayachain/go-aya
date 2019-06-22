package node

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
)

const DBPath = "/db/nodes"

type reader interface {

	GetNodeByPeerId( peerId string ) (*Node, error)

	GetFirst() *Node

	GetLatest() *Node

	TotalSum() uint64

}

type writer interface {

	Update( peerId string, node *Node ) error

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