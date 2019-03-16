package dappstate

const BroadcasterTopicPrefix  = "AyaChainBroadcaster."

//广播
type Broadcaster interface {
	OpenChannel() error
	CloseChannel() error
}

type BaseBroadcaster struct {
	Broadcaster
	state* DappState
	topics string
	Channel chan interface{}
}