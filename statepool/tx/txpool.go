package tx

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
)

const (
	TxPool_InChain 	= 0
	TxPool_Pending 	= 1
	TxPool_Queue	= 2
)

type TxPool struct {

	BaseBlock		*Block
	PendingBlock	*Block
	//由于网络的原因可能导致交易接收到的顺序不一致，此处统一把从网络中接受的防止在QueueMap中，使用PrevBlockHash作为key，并且在确定
	//是移动到Pending队列
	TxQueue			*list.List
}

func NewTxPool() (txp *TxPool) {

	txp = &TxPool{}
	txp.TxQueue = list.New()

	return
}

func (txp* TxPool) PrintTxQueue() {

	fmt.Println("\nTxPool:")
	for it := txp.TxQueue.Front(); it != nil; it = it.Next() {

		if bs, err := json.Marshal(it.Value.(Tx)); err != nil {
			log.Println(err)
		} else {
			log.Printf("TxPool:%v", string(bs))
		}

	}

}