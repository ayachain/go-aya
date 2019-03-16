package dappstate

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
	"log"
)

const BroadcasterTopicPrefix  = "AyaChainChannel."

//广播
type Broadcaster interface {
	OpenChannel() error
	CloseChannel() error
	Channel() chan interface{}
	GetTopics() string
	TypeCode() int
}

type BaseBroadcaster struct {
	Broadcaster
	typeCode int
	state* DappState
	topics string
	channel chan interface{}

	//若重写此delegate则广播会执行重写的delegate，若需要使用默认的广播方法，使此成员为nil即可
	handleDelegate func(v interface{}) error
}

func (bb *BaseBroadcaster) defaultHandleDelegate(v interface{}) error {

	var ctt string

	switch v.(type) {
	case string:
		ctt = v.(string)
	case *string:
		ctt = *(v.(*string))
	case []byte:
		ctt = hexutil.Encode(v.([]byte))
	default:
		return errors.New("Unsupport broadcaster content, the content must be string, *string, []byte. you can use json json.Marshal(v) conv content to []byte.")
	}

	return shell.NewLocalShell().PubSubPublish(bb.GetTopics(), ctt)
}

func (bb *BaseBroadcaster) Channel() chan interface{} {
	return bb.channel
}

func (bb *BaseBroadcaster) TypeCode() int {
	return bb.typeCode
}

func (bb *BaseBroadcaster) OpenChannel() error {

	go func() {

		log.Println("Broadcaster : " + bb.topics + " Channel Opened.")

		for {

			v, isopen := <- bb.channel

			if !isopen {
				log.Println("Broadcaster:" + bb.topics + "Channel Closed.")
				return
			}

			//只可能是nil 或者 error
			if bb.handleDelegate == nil {
				bb.channel <- bb.defaultHandleDelegate(v)
			} else {
				bb.channel <- bb.handleDelegate(v)
			}
		}

	}()

	return nil
}

func (bb *BaseBroadcaster) CloseChannel() error {
	close(bb.channel)
	return nil
}

func (bb *BaseBroadcaster) GetTopics() string {
	return bb.topics
}