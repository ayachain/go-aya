package dappstate

const (
	DappBroadcaster_Block 	= 1
)

func CreateBroadcaster(btype int, ds *DappState) Broadcaster {

	switch btype {
	case DappBroadcaster_Block:
		return NewBlockBroadCaseter(ds)
	default:
		return nil
	}
}