package tx

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/ayachain/go-aya/statepool/tx/act"
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
	PendingRetMap	map[string]int
	ConfirmRetCount int

	//当TxPool决定打包交易时，会通过此通道发送新块当IPFSHash，委托TxPool的拥有者进行交易广播
	//随后马上会监听此信道，查看结果，如果返回为nil，则表示成功广播
	// out chan
	BlockBroadcastChan		chan interface{}

	//矿池提供结果的信道
	// in chan
	BlockBDHashChan			chan *act.TxRspAct
}

func NewTxPool(baseBlock *Block) (txp *TxPool) {

	txp = &TxPool{}
	txp.BaseBlock = baseBlock
	txp.TxQueue = list.New()
	//测试,暂时为1
	txp.ConfirmRetCount = 1

	txp.BlockBDHashChan = make(chan *act.TxRspAct, 16)

	return
}

func (txp* TxPool) PushTx(mtx Tx) error {

	if !mtx.VerifySign() {
		return errors.New("Tx Verify Faild.")
	}

	txp.TxQueue.PushBack(mtx)

	//log.Println("TxPool Pushed New Tx : " + mtx.MarshalJson())

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
func (txp *TxPool) StartGenBlockDaemon() {

	go func() {

		for {

			//必须在PengdingBlock = nil 到情况下才可以打包新块
			if txp.PendingBlock != nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

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

				//2.创建Block,由于新块一定是Pending块，所以不会存在结果的BDHash
				//其他节点收到此块若发现BDHash长度为空，则会开始计算此块的结果
				pblock := NewBlock( txp.BaseBlock.Index + 1, txp.BaseBlock.Hash, ptxs, "")

				if bhash, err := pblock.GetHash(); err != nil {
					//若在写入时候出现问题，则交易的备份直接还原到TxQueue中
					txp.TxQueue.PushFrontList(btxList)
					log.Println(err)
					continue
				} else {
					//写入成功先广播Block，广播成功后删除备份继续等待
					//委托DappState的广播者进行交易广播
					txp.BlockBroadcastChan <- bhash
					ret := <-txp.BlockBroadcastChan

					if ret == nil {
						//广播成功
						txp.PendingBlock = pblock
						btxList.Init()

						//清除之前的所有回执
						for k, v := range txp.PendingRetMap {
							log.Println(k, v)
							delete(txp.PendingRetMap,k)
						}

						continue

					} else {

						//广播失败，还原队列
						txp.TxQueue.PushFrontList(btxList)
						btxList.Init()

						switch ret.(type) {
						case error:
							fmt.Println(ret.(error))
						}

						continue
					}
				}

			} else {
				//若无交易休眠100毫秒继续循环
				time.Sleep(time.Millisecond * 100)
			}

		}

	}()

	//监听网络上广播的计算结果
	go func() {

		for {

			rsp := <- txp.BlockBDHashChan

			if rsp.BlockHash == txp.PendingBlock.Hash {
				txp.PendingRetMap[rsp.BlockHash]++
			}

			//检测哪一个结果到数量高于x
			for k,v := range txp.PendingRetMap {

				if v > txp.ConfirmRetCount {

					//用v作为正确结果出块
					txp.PendingBlock.BDHash = k
					txp.BlockBroadcastChan <- txp.PendingBlock
					ret := <- txp.BlockBroadcastChan

					if ret == nil {
						txp.BaseBlock = txp.PendingBlock
						txp.PendingBlock = nil

					} else {
						panic(ret)
					}

				}

			}
		}

	}()

}