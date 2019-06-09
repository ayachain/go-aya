package headers

import	 (
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
)

const (
	DBPATH = "/db/header"
	latestHeaderNumKey = "LATESTHeader"
)


type HeadersAPI interface {

	AVdbComm.VDBSerices

	HeaderOf( index uint64 ) (*Header, error)

	LatestHeaderIndex() uint64

	AppendHeaders( group *AWork.TaskBatchGroup, header... *Header) error
}