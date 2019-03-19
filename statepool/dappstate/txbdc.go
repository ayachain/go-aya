package dappstate

type TxBroadCaseter struct {
	BaseBroadcaster
}

func NewTxBroadCaseter(ds* DappState) Broadcaster {

	bbc := &BlockBroadCaseter{}
	bbc.state = ds
	bbc.typeCode = PubsubChannel_Tx
	bbc.topics = BroadcasterTopicPrefix + ds.DappNS + ".Tx.Broadcast"
	bbc.channel = make(chan interface{})

	return bbc
}