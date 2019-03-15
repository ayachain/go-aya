package dappstate

import (
	TX "../tx"
	"github.com/ipfs/go-ipfs-api"
	"log"
)

type TxListener struct {
	baseListner
}

func NewTxListner( ds* DappState ) Listener {

	topics := ListnerTopicPrefix + ds.IPNSHash + ".Tx.Commit"

	newListner := &TxListener{
		baseListner:baseListner{
			state:ds,
			topics:topics,
			threadstate:ListennerThread_Stop,
		},
	}

	newListner.Listener = newListner

	return newListner
}

//收到新的交易
//func (l* baseListner) Handle(msg *shell.Message) {
func (l *TxListener) Handle ( msg *shell.Message ) {

	mtx := TX.Tx{}

	//解码返回Tx对象
	if err := mtx.DecodeFromHex(string(msg.Data)); err != nil {
		log.Print(err)
		return
	}

	//验证签名
	if !mtx.VerifySign() {
		log.Println("Tx verify faild.")
		return
	}

	//放入队列中，等待打包
	l.state.Pool.TxQueue.PushBack(mtx)

	l.state.Pool.PrintTxQueue()

}