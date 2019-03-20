package dappstate

import (
	MiningPool "github.com/ayachain/go-aya/avm"
	Miner "github.com/ayachain/go-aya/avm/miner"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	"github.com/ipfs/go-ipfs-api"
)

type BlockListener struct {

	BaseListner

	RecvBlockChan	chan *Atx.Block
}

func NewBlockListner( ds* DappState ) Listener {

	topics := BroadcasterTopicPrefix + ds.DappNS + ".Block.Broadcast"

	newListner := &BlockListener{
		BaseListner{
			state:ds,
			topics:topics,
			threadstate:ListennerThread_Stop,
		},make(chan *Atx.Block),
	}

	newListner.handleDelegate = newListner.Handle

	return newListner
}

func (l *BlockListener) Handle(msg *shell.Message) {

	//if !l.state.ContainMaterPeer(msg.From) {
	//	//不是可信任的主节点广播，不处理
	//	return
	//}

	if bcb, err := Atx.ReadBlock(string(msg.Data)); err == nil {

		if len(bcb.BDHash) > 0 {
			//有Hash一定是出块广播
			l.state.Pool.BaseBlock = bcb
			bcb.PrintIndent()

		} else {
			//否则一定是广播了一个等待计算的Block
			l.state.Pool.PendingBlock = bcb

			//l.state.DappNS
			MiningPool.AvmWorkstation.MinerChannel <- &Miner.MiningTask{
				DappNS:l.state.DappNS,
				PendingBlock:bcb,
				ResultChannel:l.state.GetBroadcastChannel(PubsubChannel_Rsp),
			}
		}

	}
}