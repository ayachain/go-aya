package signaturer

import (
	"context"
	"fmt"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/common"
	AMsgBlock "github.com/ayachain/go-aya/vdb/block"
	"time"
)

type Signturer struct {

	ACStep.ConsensusStep

	nextStep ACStep.ConsensusStep

	acceptChan chan *ADog.MsgFromDogs

	identifier string

}

func NewSignturer() ACStep.ConsensusStep {

	stp := &Signturer{
		identifier : "APOS-Step-1-Signature",
		acceptChan : make(chan *ADog.MsgFromDogs, APosComm.StepSignatureChanSize),
	}

	return stp
}

func (s *Signturer) StartListenAccept( ctx context.Context )() {

	go func() {

		fmt.Printf("%v Online\n", s.Identifier() )

		select {
		case dmsg := <- s.acceptChan :

			switch dmsg.Data[0] {
			case AMsgBlock.MessagePrefix:
				s.NextStep().ChannelAccept() <- dmsg

			}

		case <- ctx.Done() :
			break

		default:
			time.Sleep( time.Microsecond  * 100 )
		}

		fmt.Printf("%v consensus step close accept.", s.Identifier() )
	}()

}