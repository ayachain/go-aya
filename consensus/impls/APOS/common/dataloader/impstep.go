package dataloader

import (
	"context"
	"fmt"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	"github.com/ayachain/go-aya/vdb/block"
	"github.com/ipfs/go-cid"
	"time"
)

func (s *DataLoader) Identifier( ) string {
	return s.identifier
}

func (s *DataLoader) SetNextStep( ns ACStep.ConsensusStep ) {
	s.nextStep = ns
}

func (s *DataLoader) ChannelAccept() chan *ADog.MsgFromDogs {
	return s.acceptChan
}

func (s *DataLoader) NextStep() ACStep.ConsensusStep {
	return s.nextStep
}

func (s *DataLoader) Consensued( *ADog.MsgFromDogs ) {
	panic("nonreversible consensus expected")
}

func (s *DataLoader) StartListenAccept( ctx context.Context )() {

	go func() {

		fmt.Printf("%v Online\n", s.Identifier() )

		select {
		case dmsg := <- s.acceptChan:

			switch dmsg.Data[0] {
			case 'b':

				blk, ok := dmsg.ExtraData.(*block.Block)
				if !ok {
					dmsg.ResultDefer( ADog.FinalResult_ELV1 )
					break
				}

				// Load Txs blocks
				tbcid, err := cid.Decode(blk.Txs)
				if err != nil {
					dmsg.ResultDefer( ADog.FinalResult_ELV1 )
					break
				}

				subctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				txBlock, err := s.ind.Blocks.GetBlock( subctx, tbcid )
				if err != nil {
					dmsg.ResultDefer( ADog.FinalResult_ELV0 )
				}

				dmsg.ExtraData = map[string]interface{}{
					"Block" : blk,
					tbcid.String() : txBlock,
 				}

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