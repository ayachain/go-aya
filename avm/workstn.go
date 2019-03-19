package avm

import (
	"container/list"
	"github.com/ayachain/go-aya/avm/miner"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	"log"
	"time"
)

const MinerCountMaxLimit  =  10

type avmWorkstation struct {

	MinerChannel chan *Atx.Block
	vms			*list.List
	threadLimit	int
}

var AvmWorkstation = &avmWorkstation{
	MinerChannel:make(chan *Atx.Block, MinerCountMaxLimit),
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

				bcb := <- AvmWorkstation.MinerChannel

				//分配到了可以计算的虚拟机
				go func() {

					if r, err := nvm.StartSyncMining(bcb); err != nil {
						log.Println(err)
					} else {
						log.Println(r)
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