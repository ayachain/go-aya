package dappstate

import (
	Atx "github.com/ayachain/go-aya/statepool/tx"
	"github.com/ayachain/go-aya/statepool/tx/act"
	"github.com/ipfs/go-ipfs-api"
	"log"
)

type RspListener struct {
	BaseListner
	RspActOutChan			chan *act.TxRspAct
}

func NewRspListner( ds* DappState ) Listener {

	topics := BroadcasterTopicPrefix + ds.DappNS + ".Block.BDHashReply"

	newListner := &RspListener{}
	newListner.BaseListner.state = ds
	newListner.BaseListner.topics = topics
	newListner.BaseListner.threadstate = ListennerThread_Stop
	newListner.handleDelegate = newListner.Handle

	return newListner
}

func (l *RspListener) Handle(msg *shell.Message) {

	mtx := Atx.Tx{}

	//解码返回Tx对象
	if err := mtx.DecodeFromHex(string(msg.Data)); err != nil {
		log.Print(err)
		return
	}

	if !mtx.VerifySign() {
		return
	}

	rsp := &act.TxRspAct{}

	if rsp.DecodeFromHex(mtx.ActHex) != nil {
		return
	}

	l.RspActOutChan <- rsp
}
