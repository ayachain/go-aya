package dappstate

const (
	PubsubChannel_Block 	= 1
	PubsubChannel_Tx		= 2
)

func CreateBroadcaster(btype int, ds *DappState) Broadcaster {

	switch btype {
	case PubsubChannel_Block:
		return NewBlockBroadCaseter(ds)
	case PubsubChannel_Tx:
		return NewTxBroadCaseter(ds)
	default:
		return nil
	}
}