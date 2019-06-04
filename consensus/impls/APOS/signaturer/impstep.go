package signaturer

import (
	"context"
	"fmt"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	"time"
)

func (s *Signturer) Identifier( ) string {
	return s.identifier
}

func (s *Signturer) SetNextStep( ns ACStep.ConsensusStep ) {
	s.nextStep = ns
}

func (s *Signturer) ChannelAccept() chan *ADog.MsgFromDogs {
	return s.acceptChan
}

func (s *Signturer) NextStep() ACStep.ConsensusStep {
	return s.nextStep
}

func (s *Signturer) Consensued( *ADog.MsgFromDogs ) interface{} {
	return nil
}

func (s *Signturer) StartListenAccept( ctx context.Context )() {

	go func() {

		fmt.Printf("%v consensus step start accept.", s.Identifier() )

		select {
		case dmsg := <- s.acceptChan :

			//Next Step
			s.NextStep().ChannelAccept() <- dmsg

			//Reject
			dmsg.ResultDefer( ADog.FinalResult_ELV1 )

		case <- ctx.Done() :
			break

		default:
			time.Sleep( time.Microsecond  * 100 )
		}

		fmt.Printf("%v consensus step close accept.", s.Identifier() )
	}()

}