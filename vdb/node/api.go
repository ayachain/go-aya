package node

import (
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/indexes"
)

const DBPath = "/nodes"

type reader interface {

	GetNodeByPeerId( peerId string, idx... *indexes.Index ) (*im.Node, error)

	GetSuperMaterTotalVotes( idx... *indexes.Index ) uint64

	GetSuperNodeList( idx... *indexes.Index ) []*im.Node

	GetSuperNodeCount( idx... *indexes.Index ) int64

	DoRead( readingFunc ADB.ReadingFunc, idx... *indexes.Index ) error
}

type writer interface {

	InsertBootstrapNodes( nds []*im.Node )

	Insert( peerId string, node *im.Node ) error

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