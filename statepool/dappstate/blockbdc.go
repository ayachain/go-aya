package dappstate

type BlockBroadCaseter struct {
	BaseBroadcaster
}

func NewBlockBroadCaseter(ds* DappState) Broadcaster {

	bbc := &BlockBroadCaseter{}
	bbc.state = ds
	bbc.topics = BroadcasterTopicPrefix + ds.IPNSHash + ".Block.Broadcast"
	bbc.Channel = make(chan interface{})

	return bbc
}

func (bb *BlockBroadCaseter) OpenChannel() error {
	return nil
}

func (bb *BlockBroadCaseter) CloseChannel() error {
	return nil
}