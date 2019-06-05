package worker

import (
	"context"
	"encoding/json"
	"fmt"
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ayachain/go-aya/vdb"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-ipfs/core"
	"time"
)

func (s *Worker) Identifier( ) string {
	return s.identifier
}

func (s *Worker) SetNextStep( ns ACStep.ConsensusStep ) {
	s.nextStep = ns
}

func (s *Worker) ChannelAccept() chan *ADog.MsgFromDogs {
	return s.acceptChan
}

func (s *Worker) NextStep() ACStep.ConsensusStep {
	return s.nextStep
}

func (s *Worker) Consensued( *ADog.MsgFromDogs ) interface{} {
	return nil
}

func (s *Worker) StartListenAccept( ctx context.Context )() {

	go func() {

		fmt.Printf("%v consensus step start accept.", s.Identifier() )

		select {
		case dmsg := <- s.acceptChan :

			switch dmsg.Data[0] {
			case 'b':

				emap, ok := dmsg.ExtraData.(map[string]interface{})
				if !ok {
					dmsg.ResultDefer(ADog.FinalResult_ELV0)
					break
				}

				blk, ok := emap["Block"].(*ABlock.Block)
				if !ok {
					dmsg.ResultDefer(ADog.FinalResult_ELV0)
					break
				}

				txsBlock, ok := emap[blk.Txs].(blocks.Block)
				if !ok {
					dmsg.ResultDefer(ADog.FinalResult_ELV0)
					break
				}

				dmsg.ExtraData = doWorking( s.ind, txsBlock, blk )
				if dmsg.ExtraData == nil {
					dmsg.ResultDefer(ADog.FinalResult_ELV0)
					break
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


func doWorking( ind *core.IpfsNode, txsBlock blocks.Block, block *ABlock.Block ) *AWork.TaskBatchGroup {

	txs := make([]*ATx.Transaction, block.Txc)
	if err := json.Unmarshal(txsBlock.RawData(), txs); err != nil {
		return nil
	}

	vfs, err := vdb.CreateVFS( block, ind )
	if err != nil {
		return nil
	}

	batchgroup := AWork.NewGroup()
	for _, tx := range txs {

		if len(tx.Data) == 0 {

			err := transfer(tx, batchgroup, vfs)

			if err != nil {
				return nil
			}

		}

	}

	return batchgroup
}