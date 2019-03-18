package dappstate

import (
	Atx "github.com/ayachain/go-aya/statepool/tx"
	"github.com/ipfs/go-ipfs-api"
)

type BlockListener struct {

	BaseListner

	RecvBlockChan	chan *Atx.Block
}

func NewBlockListner( ds* DappState ) Listener {

	topics := BroadcasterTopicPrefix + ds.IPNSHash + ".Block.Broadcast"

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

	/*
	逻辑说明：在BlockListen接收到广播的块以后需要完成以下逻辑
	1.确认广播人发送方为当前Dapp中定义主节点
	2.如果本地的交易池没有BaseBlock则把接收的块当做BaseBlock然后继续等待后续的出块广播
	3.如果本地的交易池有BaseBlock但是没有PandingBlock
		3.1 如果接收到的新块是BaseBlock的下一块，则把新块当做PandingBlock
		3.2 如果接收到的新块不是BaseBlock的下一块，则把新块当做BaseBlock继续等待后续广播
	4.如果本地的交易池有正确的BaseBlock和PandingBlock
		4.1 如果接收的新块是PandingBlock的下一块则表示为出块广播此时把PendingBlock设为BaseBlock，新块作为PandingBlock
		4.2 如果为接收到正确的Block，或BaseBlock，PandingBlock和新块之间不能衔接，则把新块作为BaseBlock等待广播
	*/

	//if !l.state.ContainMaterPeer(msg.From) {
	//	//不是可信任的主节点广播，不处理
	//	return
	//}

	if bcb, err := Atx.ReadBlock(string(msg.Data)); err == nil {

		bcb.PrintIndent()

		switch bcb.Index {
		case 0 :
			//创世块
			l.state.Pool.BaseBlock = bcb

		case 1 :
			l.state.Pool.PendingBlock = bcb

		default:

			if l.state.Pool.BaseBlock == nil {
				l.state.Pool.BaseBlock = bcb
			} else if bcb.Index - l.state.Pool.BaseBlock.Index == 1 && bcb.Prev == l.state.Pool.BaseBlock.Hash {
				//当收到的新块是BaseBlock的下一块并且在无PengdingBlock 的情况下，当节点认为此块为PandingBlock
				//因为在缺失BaseBlock和PanedingBlock的时候，节点是不能够处理交易数据的
				l.state.Pool.PendingBlock = bcb
			} else if bcb.Index - l.state.Pool.PendingBlock.Index == 1 && bcb.Prev == l.state.Pool.PendingBlock.Hash {
				//当新块是Pengding当下一块，表示主节点已经决定出块,那么本地节点应该马上放弃PengdingBlock当计算，去计算最新当PangdingBLock
				l.state.Pool.BaseBlock = l.state.Pool.PendingBlock
				l.state.Pool.PendingBlock = bcb
			} else {
				//其他情况，有如下可能
				//1.本节点刚刚启动，还为获取到任何数据，所以BaseBlock和PendingBlock都没有数据
				//2.网络延迟原因，导致中间丢失了一块都广播，导致和本地交易池无法正确连接
				//3.收到了错误的Block那么
				//上述情况会本节点认定为暂时不可以计算交易，则需要重置BaseBlock和PendingBlock，等待主节点至少广播连续的两块后才可以开始计算
				l.state.Pool.BaseBlock = bcb
				l.state.Pool.PendingBlock = nil
			}
		}
	}
}