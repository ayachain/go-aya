package tx

const (
	TxPool_InChain 	= 0
	TxPool_Pending 	= 1
	TxPool_Queue	= 2
)

type TxPool struct {
	BaseChain 		chain
	PendingChain 	memChain

	//由于网络的原因可能导致交易接收到的顺序不一致，此处统一把从网络中接受的防止在QueueMap中，使用PrevBlockHash作为key，并且在确定
	//是移动到Pending队列
	QueueMap		map[string]*Block

}

func (txp *TxPool) SortPool() {

	//检测是否存在于暂未确定的队列中
	for {

		//PendingChain last block hash
		plbh, err := txp.PendingChain.LatestBlock.GetHash()

		if err != nil{
			return
		}

		_, exist := txp.QueueMap[plbh]

		if exist {

			if err := txp.PendingChain.AppendBlock(txp.QueueMap[plbh]); err != nil {
				return
			} else {
				//移动Block成功删除QueueMap中对应的Block
				delete(txp.QueueMap, plbh)
			}

		} else {
			break
		}
	}

}


func (txp *TxPool) PushBlock(b* Block) {
	txp.QueueMap[b.Prev] = b
}