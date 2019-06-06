package signaturer

import (
	"context"
	"fmt"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/common"
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

			switch dmsg.Data[0] {
			case 'b':

				// If it is a message of block data, it goes directly to the next step,
				// because block data is unsigned and only the source of the data provides
				// the judgment.
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