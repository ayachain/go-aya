package dappstate

type BlockBroadCaseter struct {
	BaseBroadcaster
}

func NewBlockBroadCaseter(ds* DappState) Broadcaster {

	bbc := &BlockBroadCaseter{}
	bbc.state = ds
	bbc.typeCode = PubsubChannel_Block
	bbc.topics = BroadcasterTopicPrefix + ds.DappNS + ".Block.Broadcast"
	bbc.channel = make(chan interface{})

	return bbc
}