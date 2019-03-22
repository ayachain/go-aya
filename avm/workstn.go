package avm

import (
	"container/list"
	"github.com/ayachain/go-aya/avm/miner"
	Avmm "github.com/ayachain/go-aya/avm/miner/module"
	"log"
	"time"
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

			//查找已经生成的空闲的虚拟机
			for it := AvmWorkstation.vms.Front(); it != nil; it = it.Next() {
				if it.Value.(*miner.Avm).State == miner.AvmState_IDLE {
					nvm = it.Value.(*miner.Avm)
				}
			}

			if nvm == nil {

				//若无空闲虚拟机，查看是否到达配置的虚拟机最大上线
				if AvmWorkstation.vms.Len() < AvmWorkstation.threadLimit {
					//未到达最大限制
					nvm = miner.NewAvm()
					AvmWorkstation.vms.PushBack(nvm)
				}

			}

			if nvm != nil {

				task := <- AvmWorkstation.MinerChannel

				//分配到了可以计算的虚拟机
				go func() {

					Avmm.SetAvmBasePath(nvm.GetL(), "/" + task.PendingBlock.GetHash() + "/_data")

					if r, err := nvm.StartSyncMining(task.PendingBlock); err != nil {
						log.Println(err)
					} else {
						task.ResultChannel <- task.CreateResult(r)
						bbcRet := <- task.ResultChannel

						if bbcRet != nil {
							//未能成功广播结果，丢弃
							log.Println(bbcRet.(error))
						}
					}

				}()

				continue

			} else {
				//未能成功分配到可以用于计算的虚拟机, 说明所有虚拟机都处于忙碌状态等待
				time.Sleep(time.Millisecond * 200)
			}
		}

	}()

}