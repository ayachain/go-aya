package dappstate

type RspBroadCaseter struct {
	BaseBroadcaster
}

func NewRspBroadCaseter(ds* DappState) Broadcaster {

	bbc := &BlockBroadCaseter{}
	bbc.state = ds
	bbc.typeCode = PubsubChannel_Rsp
	bbc.topics = BroadcasterTopicPrefix + ds.DappNS + ".Block.BDHashReply"
	bbc.channel = make(chan interface{})

	return bbc
}