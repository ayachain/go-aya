package dappstate

import (
	Atx "github.com/ayachain/go-aya/statepool/tx"
	"github.com/ipfs/go-ipfs-api"
	"log"
)

type TxListener struct {
	BaseListner
	handTxCount uint64
}

func NewTxListner( ds* DappState ) Listener {

	topics := BroadcasterTopicPrefix + ds.IPNSHash + ".Tx.Commit"

	newListner := &TxListener{}
	newListner.BaseListner.state = ds
	newListner.BaseListner.topics = topics
	newListner.BaseListner.threadstate = ListennerThread_Stop
	newListner.handleDelegate = newListner.Handle
	newListner.handTxCount = 0

	return newListner
}

func (l *TxListener) Handle(msg *shell.Message) {

	mtx := Atx.Tx{}

	//解码返回Tx对象
	if err := mtx.DecodeFromHex(string(msg.Data)); err != nil {
		log.Print(err)
		return
	}

	//放入队列中，等待打包
	if err := l.state.Pool.PushTx(mtx); err != nil {
		log.Println(err)
		return
	}
}
