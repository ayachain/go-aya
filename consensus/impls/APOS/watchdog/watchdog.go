package watchdog

import (
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"

	"github.com/libp2p/go-libp2p-pubsub"
)

type Dog struct {
	ADog.WatchDog
	cc *ACStep.ConsensusChain
	nextStepChan chan *ADog.MsgFromDogs
}

func NewDog( cc *ACStep.ConsensusChain ) *Dog {
	return &Dog{
		nextStepChan:cc.ChannelAccept(),
	}
}

func (d *Dog) TakeMessage( msg pubsub.Message ) *ADog.MsgFromDogs {

	rfunc := func( ADog.FinalResult )  {
		//Waiting dev
	}

	dmsg := &ADog.MsgFromDogs{
		Message:msg,
		ResultDefer:rfunc,
	}

	d.nextStepChan <- dmsg

	return nil
}

func (d *Dog) CreditScoring( peeridOrAddress string ) int8 {
	return 100
}

func (d *Dog) Close() {

}
