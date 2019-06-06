package watchdog

import (
	"context"
	"errors"
	"fmt"
	AMsgBlock "github.com/ayachain/go-aya/chain/message/block"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/common"
	"github.com/libp2p/go-libp2p-pubsub"
	"time"
)

type Dog struct {

	ADog.WatchDog
	cc *ACStep.ConsensusChain

	rules map[byte]*ACStep.ConsensusChain
	recvChan chan *ADog.MsgFromDogs
}

func NewDog( ) *Dog {

	return &Dog{
		rules		: make(map[byte]*ACStep.ConsensusChain),
		recvChan	: make(chan *ADog.MsgFromDogs, APosComm.StepWatchDogChanSize),
	}

}


func (d *Dog) SetRule( msgtype byte, cc *ACStep.ConsensusChain ) {
	d.rules[msgtype] = cc
}


func (d *Dog) TakeMessage( msg *pubsub.Message ) error {

	rfunc := func( ADog.FinalResult )  {
		//Waiting dev
	}

	switch msg.Data[0] {
	case AMsgBlock.MessagePrefix :

		blk := AMsgBlock.MsgRawBlock{}
		if err := blk.Decode(msg.Data[1:]); err != nil {
			return err
		}

		d.recvChan <- &ADog.MsgFromDogs{
			Message:msg,
			ExtraData:blk,
			ResultDefer:rfunc,
		}

		return nil

	default:
		fmt.Println("unkown message type")
	}

	return errors.New("unkown message type")
}

func (d *Dog) CreditScoring( peeridOrAddress string ) int8 {
	return 100
}

func (d *Dog) Close() {

}


func (d *Dog) StartListenAccept( ctx context.Context ) {

	go func() {

		select {

		// AI Handle
		// Wait dev
		case msg := <- d.recvChan : {
			cc, exist := d.rules[msg.Data[0]]
			if !exist {
				break
			}

			cc.GetStepRoot().ChannelAccept() <- msg
		}

		case <- ctx.Done(): {
			return
		}

		default:
			time.Sleep( time.Microsecond  * 100 )
		}

	}()

}