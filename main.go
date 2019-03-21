package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ayachain/go-aya/avm"
	Aks "github.com/ayachain/go-aya/keystore"
	DState "github.com/ayachain/go-aya/statepool/dappstate"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	"log"
	"time"
)

const (
	AyaChainDemoDapp_1 = "QmP5RqvBkfW6NhA6h3rajd71maWm7pUSbVyk9syxdk856h"
	AyaChainDemoDapp_2 = "QmcCAXw29EcssMiLvF4WhMDY5nLzqv6AxaZB3wgRqijG8c"
)

func main() {

	nodeType := flag.String("t", "worker","")

	flag.Parse()

	peerType := DState.DappPeerType_Worker

	if *nodeType == "master" {
		peerType = DState.DappPeerType_Master
	}

	//生成一个Dapp状态机
	fristDemoState, err := DState.NewDappState(AyaChainDemoDapp_2)

	if err != nil {
		panic(err)
	}

	//启动状态机守护线程，当中有主题的监听
	if err := fristDemoState.Daemon(peerType); err == nil {

		log.Println("AyaChain Daemon is ready.")

		//启动测试线程发送交易
		go func() {

			txindex := 0

			for {

				txindex++

				//内部休眠时间为100毫秒 所以在保证在100毫秒内可以发送3比交易测试
				time.Sleep(time.Millisecond * 100)
				//time.Sleep(time.Second * 2)

				pmap := make(map[string]string)
				pmap["name"] = fmt.Sprintf("[VisiterNumber:%d]", txindex)
				pmapbs, _ := json.Marshal(pmap)

				var act Act.BaseActInf

				switch txindex % 3 {
				case 0 :
					act = Act.NewPerfromAct(AyaChainDemoDapp_2, "SayHello", string(pmapbs))

				case 1 :
					act = Act.NewPerfromAct(AyaChainDemoDapp_2, "GiveMeALTable", "")

				case 2 :
					act = Act.NewPerfromAct(AyaChainDemoDapp_2, "GiveMeANumber", "")
				}



				//签名
				tx := Aks.DefaultPeerKS().CreateSignedTx(act)

				if txhex, err := tx.EncodeToHex(); err == nil {

					fristDemoState.GetBroadcastChannel(DState.PubsubChannel_Tx) <- txhex
					//获取结果
					ret := <- fristDemoState.GetBroadcastChannel(DState.PubsubChannel_Tx)

					if ret != nil {
						panic(err)
					}

				} else {
					panic(err)
				}
			}
		}()

		avm.DaemonWorkstation()
		//阻塞主线程
		select {}

	} else {

		panic(err)

	}
}