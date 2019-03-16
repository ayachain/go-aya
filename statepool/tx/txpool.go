package tx

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"time"
)

const (
	TxPool_InChain 	= 0
	TxPool_Pending 	= 1
	TxPool_Queue	= 2
)

type TxPool struct {

	BaseBlock		*Block
	PendingBlock	*Block
	TxQueue			*list.List

	//当TxPool决定打包交易时，会通过此通道发送新块当IPFSHash，委托TxPool的拥有者进行交易广播
	//随后马上会监听此信道，查看结果，如果返回为nil，则表示成功广播
	BlockBroadcastChan		chan string
}

func NewTxPool() (txp *TxPool) {

	txp = &TxPool{}
	txp.TxQueue = list.New()
	txp.BlockBroadcastChan = make(chan string)

	return
}

func (txp* TxPool) PushTx(mtx Tx) error {

	if !mtx.VerifySign() {
		return errors.New("Tx Verify Faild.")
	}

	txp.TxQueue.PushBack(mtx)

	log.Println("TxPool Pushed New Tx : " + mtx.MarshalJson())

	return nil
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

//交易打包线程
//1.若交易池中无交易则直接休眠100毫秒，为了保证相应的即时性此处的休眠时间可以根据情况跳转，暂时设置在100毫秒
//2.若交易池中有滞留的交易，并且上一块已经确定，则直接打包Block进行广播
//3.若交易池中有滞留的交易，但是上一块还为确定，则需要等待上一块确认
func (txp *TxPool) GenBlockDaemon() {

	go func() {

		for {

			if txp.TxQueue.Len() > 0 {

				//记录当前队列的数量（由于外部线程会持续写入交易，但因为此处是队列存放，所以需要记录开始打包交易时候，收到的数量，防止在处理过程中，被追加交易导致异常）
				//每个Block最多保存1024个交易，超出部分则等待下一块
				var ptxcount int

				if txp.TxQueue.Len() > 1024 {
					ptxcount = 1024
				} else {
					ptxcount = txp.TxQueue.Len()
				}

				//打包Block
				//1.交易处理，删除内存队列时候，必须保证已经正确的上传到了IPFS
				btxList := list.New()
				ptxs := make([]Tx, ptxcount)

				for i := 0; i < ptxcount; i++ {

					//从队列的起始位置依次移动交易进行打包
					fit := txp.TxQueue.Front()
					ptxs[i] = fit.Value.(Tx)

					//备份
					btxList.PushBack(fit.Value)
					txp.TxQueue.Remove(fit)
				}

				//2.创建Block
				pblock := NewBlock( txp.PendingBlock.Index + 1, txp.PendingBlock.Hash, ptxs, "")

				if bhash, err := pblock.WriteBlock(); err != nil {
					//若在写入时候出现问题，则交易的备份直接还原到TxQueue中
					txp.TxQueue.PushFrontList(btxList)
					log.Println(err)
					continue
				} else {
					//写入成功先广播Block，广播成功后删除备份继续等待
					//委托DappState进行交易广播
					txp.BlockBroadcastChan <- bhash
					ret := <-txp.BlockBroadcastChan

					if ret == "" {
						//广播成功
						btxList.Init()
						continue

					} else {
						//广播失败
						txp.TxQueue.PushFrontList(btxList)
						btxList.Init()
						continue
					}
				}


			} else {
				//若无交易休眠500毫秒继续循环
				time.Sleep(time.Millisecond * 100)
			}

		}

	}()

}