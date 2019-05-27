package avm

import (
	"container/list"
	"github.com/ayachain/go-aya/avm/miner"
	"log"
)

const MinerCountMaxLimit = 128

type avmWorkstation struct {
	MinerChannel chan *miner.MiningTask
	vms			*list.List
	threadLimit	int
}

var AvmWorkstation = &avmWorkstation{
	MinerChannel:make(chan *miner.MiningTask, MinerCountMaxLimit),
	vms:list.New(),
	threadLimit:MinerCountMaxLimit,
}

func DaemonWorkstation() {

	go func() {

		for {

			var nvm *miner.Avm

			////查找已经生成的空闲的虚拟机
			//for it := AvmWorkstation.vms.Front(); it != nil; it = it.Next() {
			//
			//	if it.Value.(*miner.Avm).State == miner.AvmState_IDLE {
			//		nvm = it.Value.(*miner.Avm)
			//	}
			//
			//}

			//若无空闲虚拟机，查看是否到达配置的虚拟机最大上线
			if AvmWorkstation.vms.Len() < AvmWorkstation.threadLimit {

				//未到达最大限制
				task := <- AvmWorkstation.MinerChannel
				nvm = miner.NewAvm(task.DappNS)
				AvmWorkstation.vms.PushBack(nvm)

				//分配到了可以计算的虚拟机
				go func() {

					if r, err := nvm.StartSyncMining(task.PendingBlock); err != nil {
						log.Println(err)
					} else {
						task.ResultChannel <- task.CreateResult(r)
						bbcRet := <-task.ResultChannel

						if bbcRet != nil {
							//未能成功广播结果，丢弃
							log.Println(bbcRet.(error))
						}
					}

				}()
			}

		}
	}()

}