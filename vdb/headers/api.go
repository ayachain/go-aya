package headers


const (
	headersDBPath = "/db/header"
	latestHeaderNumKey = "LATESTHeader"
)


type HeadersAPI interface {

	DBKey()	string

	HeaderOf( index uint64 ) (*Header, error)

	LatestHeaderIndex() uint64

	AppendHeaders( header... *Header) error
}