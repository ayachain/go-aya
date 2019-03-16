package dappstate

import (
	TX "../tx"
	"github.com/ipfs/go-ipfs-api"
	"strings"
)

type BlockListener struct {

	BaseListner

	RecvBlockChan	chan *TX.Block

}

func NewBlockListner( ds* DappState ) Listener {

	topics := ListnerTopicPrefix + ds.IPNSHash + ".Block.Broadcast"

	newListner := &BlockListener{
		BaseListner{
			state:ds,
			topics:topics,
			threadstate:ListennerThread_Stop,
		},make(chan *TX.Block),
	}

	//newListner.Listener = newListner

	return newListner
}

func (l *BlockListener) Handle(msg *shell.Message) {

	/*
	逻辑总体说明：在BlockListen接收到广播的块以后需要完成以下逻辑
	1.确认广播人发送方为当前Dapp中定义主节点
	2.如果本地的交易池没有BaseBlock则把接收的块当做BaseBlock然后继续等待后续的出块广播
	3.如果本地的交易池有BaseBlock但是没有PandingBlock
		3.1 如果接收到的新块是BaseBlock的下一块，则把新块当做PandingBlock
		3.2 如果接收到的新块不是BaseBlock的下一块，则把新块当做BaseBlock继续等待后续广播
	4.如果本地的交易池有正确的BaseBlock和PandingBlock
		4.1 如果接收的新块是PandingBlock的下一块则表示为出块广播此时把PendingBlock设为BaseBlock，新块作为PandingBlock
		4.2 如果为接收到正确的Block，或BaseBlock，PandingBlock和新块之间不能衔接，则把新块作为BaseBlock等待广播
	*/

	if !l.state.ContainMaterPeer(msg.From) {
		//不是可信任的主节点广播，不处理
		return
	}

	if bcb, err := TX.ReadBlock(string(msg.Data)); err == nil {

		if l.state.Pool.BaseBlock == nil {
			//若本节点交易池没有任何数据，接收到新块直接认为是已确认到块
			l.state.Pool.BaseBlock = bcb
			return
		}

		baseHash, _ := l.state.Pool.BaseBlock.GetHash()

		if l.state.Pool.PendingBlock == nil {

			//当前节点有确认到Block记录，但是没有正在Pending的记录
			if strings.EqualFold(bcb.Prev,baseHash) {

				//接收的新块正好是Pending块
				l.state.Pool.PendingBlock = bcb

				//正确接收Pending块后，通过管道将PandingBlock发送到外部
				l.RecvBlockChan <- bcb

				return

			} else {

				//若接收的新块不能和Base连接，则把当前新块当做BaseBlock,然后继续等待,Pending块的广播
				l.state.Pool.BaseBlock = bcb

				return
			}

		} else {

			pendingHash, _ := l.state.Pool.PendingBlock.GetHash()

			if strings.EqualFold(bcb.Prev, pendingHash) {

				//接收的新块正好是Pending的下一块，表示出块
				l.state.Pool.BaseBlock = l.state.Pool.PendingBlock
				l.state.Pool.PendingBlock = bcb

				return

			} else {

				//如果无法衔接，则尝试匹配是否又接收到了PendingBlock
				if strings.EqualFold(bcb.Prev, baseHash) {

					l.state.Pool.PendingBlock = bcb
					return

				} else {

					l.state.Pool.BaseBlock = bcb
					l.state.Pool.PendingBlock = nil
					return

				}
			}
		}
	}
}