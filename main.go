package main

import (
	"flag"
	"github.com/ayachain/go-aya/avm"
	Aks "github.com/ayachain/go-aya/keystore"
	DState "github.com/ayachain/go-aya/statepool/dappstate"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	"log"
	"strconv"
	"time"
)

func main() {

	nodeType := flag.String("t", "worker","")

	flag.Parse()

	peerType := DState.DappPeerType_Worker

	if *nodeType == "master" {
		peerType = DState.DappPeerType_Master
	}

	//生成一个Dapp状态机
	fristDemoState, err := DState.NewDappState("QmcBx4Ua8WmZPE9At81jRiAnjBYviD7V8noGtG5teEQTnh")

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
				//time.Sleep(time.Millisecond * 35)
				time.Sleep(time.Second * 2)

				act := Act.NewPerfromAct("QmcBx4Ua8WmZPE9At81jRiAnjBYviD7V8noGtG5teEQTnh", "main", []string{"Parmas1", strconv.Itoa(txindex)})

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