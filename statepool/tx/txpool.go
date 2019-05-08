package tx

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	"github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"time"
)

const (
	TxState_NotFound = 0
	TxState_WaitPack = 1
	TxState_Pending  = 2
	TxState_Confirm  = 3
)

type TxPool struct {

	DappNS			string
	BaseBlock		*Block
	PendingBlock	*Block
	TxQueue			*list.List

	ConfirmRetCount int

	//当TxPool决定打包交易时，会通过此通道发送新块当IPFSHash，委托TxPool的拥有者进行交易广播
	//随后马上会监听此信道，查看结果，如果返回为nil，则表示成功广播
	// out chan
	BlockBroadcastChan	chan interface{}

	//矿池提供结果的信道
	// in chan
	BlockBDHashChan		chan *Tx

	pendingRetMap		map[string]*list.List

}

func NewTxPool(dappns string) (txp *TxPool) {

	txp = &TxPool{}
	txp.TxQueue = list.New()

	//测试,暂时为1
	txp.ConfirmRetCount = 1
	txp.BlockBDHashChan = make(chan *Tx)
	txp.pendingRetMap = make(map[string]*list.List)
	txp.DappNS = dappns

	ipfssh := shell.NewLocalShell()

	objstat, err := ipfssh.ObjectStat("/ipns/" + dappns + "/_index/_bindex")

	if err != nil {
		return nil
	}

	if objstat.BlockSize < 46 {

		//块检索有数据，说明不是首次运行的Dapp直接读取最后一个记录的Block作为BaseBlock
		txp.BaseBlock = NewBlock(0,"",nil, dappns)

	} else {

		//读出当前记录完成的最后一块
		req, err := ipfssh.Request("cat").Option("o", objstat.BlockSize - 46).Option("l",46).Send(context.Background())

		if err != nil {
			return nil
		}

		if reqbs, err := ioutil.ReadAll(req.Output); err != nil {

			return nil

		} else {

			baseBlockHash := string(reqbs)

			if blkbs, err := ipfssh.BlockGet(baseBlockHash); err != nil {
				return nil
			} else {
				txp.BaseBlock = &Block{}
				if json.Unmarshal(blkbs, txp.BaseBlock) != nil {
					return nil
				}
			}

		}

	}

	return
}

func (txp* TxPool) PushTx(mtx Tx) error {

	if !mtx.VerifySign() {
		return errors.New("Tx Verify Faild.")
	}

	txp.TxQueue.PushBack(mtx)

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

	//打包线程
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
				nblock := NewBlock( txp.BaseBlock.Index + 1, txp.BaseBlock.GetHash(), ptxs, "")

				if bhash := nblock.GetHash(); len(bhash) <= 0 {

					log.Println("TxPool: Writing new block to ipfs faild. continue and retry at next loop.")

					//若在写入时候出现问题，则交易的备份直接还原到TxQueue中
					txp.TxQueue.PushFrontList(btxList)

					continue

				} else {

					//写入成功先广播Block，广播成功后删除备份继续等待
					//委托DappState的广播者进行交易广播
					txp.BlockBroadcastChan <- bhash
					ret := <-txp.BlockBroadcastChan

					if ret == nil {

						//广播成功
						txp.PendingBlock = nblock
						btxList.Init()

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
	//出块线程
	go func() {

		for {

			rsptx := <- txp.BlockBDHashChan
			rsp := &Act.TxRspAct{}

			if rsp.DecodeFromHex(rsptx.ActHex) != nil {
				return
			}

			if rsp.BlockHash == txp.PendingBlock.GetHash() {

				_,exist := txp.pendingRetMap[rsp.ResultState]

				if !exist {
					txp.pendingRetMap[rsp.ResultState] = list.New()
				}

				txp.pendingRetMap[rsp.ResultState].PushBack(rsptx)
			}

			//检测哪一个结果到数量高于x
			for k,v := range txp.pendingRetMap {

				if v.Len() >= txp.ConfirmRetCount {

					//用k作为正确结果出块
					txp.PendingBlock.BDHash = k
					txp.BlockBroadcastChan <- txp.PendingBlock.RefreshHash()
					ret := <- txp.BlockBroadcastChan

					if ret == nil {

						//出块广播成功，清除之前的所有回执更新本机对应的IPNS
						txp.pendingRetMap = make(map[string]*list.List)
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

func (txp *TxPool) SearchTxStatus(txhash string) (bindex uint64, tx *Tx, stat int, receipt *TxReceipt) {

	//if in waiting package status
	for it := txp.TxQueue.Front(); it != nil; it = it.Next() {
		if it.Value.(*Tx).GetSha256Hash() == txhash {
			return 0, tx, TxState_WaitPack, nil
		}
	}


	//if in pending status
	if txp.PendingBlock != nil {

		for i := 0; i < len(txp.PendingBlock.Txs); i++ {
			if txp.PendingBlock.Txs[i].GetSha256Hash() == txhash {
				return txp.PendingBlock.Index, tx, TxState_Pending, nil
			}
		}

	}

	//if in confirm status
	r, err := readTxReceipt(txp.DappNS, txhash)

	if err != nil {
		return 0,nil, TxState_NotFound, nil
	} else {

		return r.BlockIndex,nil,TxState_Confirm, r
	}

}

func readTxReceipt(bpath string, txhash string) (receipt *TxReceipt, err error) {

	mfsTpath := "/" + bpath + "/_receipt/" + txhash

	bs, err := shell.NewLocalShell().Request("files/read").Arguments(mfsTpath).Send(context.Background())

	if err != nil {
		return nil, err
	} else {

		if bs.Error != nil {
			return nil, errors.New( bs.Error.Error() )
		}

		if content,err := ioutil.ReadAll(bs.Output); err != nil {
			return nil, err
		} else {

			r := &TxReceipt{}

			if err := json.Unmarshal(content, r); err != nil {
				return nil, err
			}

			return r, nil
		}
	}
}