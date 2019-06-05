package executor

import (
	"context"
	"fmt"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	"time"
)

func (s *Executor) Identifier( ) string {
	return s.identifier
}

func (s *Executor) SetNextStep( ns ACStep.ConsensusStep ) {

}

func (s *Executor) ChannelAccept() chan *ADog.MsgFromDogs {
	return s.acceptChan
}

func (s *Executor) NextStep() ACStep.ConsensusStep {
	return nil
}

func (s *Executor) Consensued( *ADog.MsgFromDogs ) interface{} {
	return nil
}

func (s *Executor) StartListenAccept( ctx context.Context )() {

	go func() {

		fmt.Printf("%v consensus step start accept.", s.Identifier() )

		select {
		case dmsg := <- s.acceptChan :

			switch dmsg.Data[0] {
			case 'b':

				fbg, ok := dmsg.ExtraData.(*AWork.TaskBatchGroup)
				if !ok {
					dmsg.ResultDefer( ADog.FinalResult_ELV0 )
				}

				if err := s.cvfs.WriteBatchGroup(fbg); err != nil {
					dmsg.ResultDefer( ADog.FinalResult_ELV0 )
				}

				dmsg.ResultDefer( ADog.FinalResult_Success )
			}

		case <- ctx.Done() :
			break

		default:
			time.Sleep( time.Microsecond  * 100 )
		}

		fmt.Printf("%v consensus step close accept.", s.Identifier() )
	}()

}