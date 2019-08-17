package node

import (
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
)

const DBPath = "/nodes"

type reader interface {

	GetNodeByPeerId( peerId string, idx... *indexes.Index ) (*Node, error)

	GetSuperMaterTotalVotes( idx... *indexes.Index ) uint64

	GetSuperNodeList( idx... *indexes.Index ) []*Node

	GetSuperNodeCount( idx... *indexes.Index ) int64

	DoRead( readingFunc ADB.ReadingFunc, idx... *indexes.Index ) error
}

type writer interface {

	InsertBootstrapNodes( nds []Node )

	Insert( peerId string, node *Node ) error

	Del( peerId string )
}


type Services interface {
	AVdbComm.VDBSerices
	reader
}


type MergeWriter interface {
	AVdbComm.VDBCacheServices
	reader
	writer
}