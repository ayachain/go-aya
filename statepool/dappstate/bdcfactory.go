package dappstate

const (
	PubsubChannel_Block 	= 1
	PubsubChannel_Tx		= 2
	PubsubChannel_Rsp 		= 3
)

func CreateBroadcaster(btype int, ds *DappState) Broadcaster {

	switch btype {
	case PubsubChannel_Block:
		return NewBlockBroadCaseter(ds)
	case PubsubChannel_Tx:
		return NewTxBroadCaseter(ds)
	case PubsubChannel_Rsp:
		return NewRspBroadCaseter(ds)
	default:
		return nil
	}
}