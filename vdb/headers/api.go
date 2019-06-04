package headers


const (
	headersDBPath = "/db/header"
	latestHeaderNumKey = "LATESTHeader"
)


type HeadersAPI interface {

	HeaderOf( index uint64 ) (*Header, error)

	LatestHeaderIndex() uint64

	AppendHeaders( verify bool, header... *Header) error
}