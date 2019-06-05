package headers

import	 (
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

	AppendHeaders( header... *Header) error
}