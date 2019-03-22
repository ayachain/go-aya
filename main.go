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
	AyaChainDemoDapp_3 = "QmbLSVGXVJ4dMxBNkxneThhAnXVBWdGp7i2S42bseXh2hS"
)


func testDemo3() {

	//生成一个Dapp状态机
	fristDemoState, err := DState.NewDappState(AyaChainDemoDapp_3)

	if err != nil {
		panic(err)
	}

	//启动状态机守护线程，当中有主题的监听
	if err := fristDemoState.Daemon(DState.DappPeerType_Master); err != nil {
		panic(err)
	}

	avm.DaemonWorkstation()

	//开始交易
	//1.申请一些代币
	parmas := make(map[string]interface{})
	parmas["address"] = Aks.DefaultPeerKS().Address()
	parmas["amount"] = 5000

	jbs, _ := json.Marshal(parmas)
	act := Act.NewPerfromAct(AyaChainDemoDapp_3, "giveMeSomeToken", string(jbs))

	//签名
	tx := Aks.DefaultPeerKS().CreateSignedTx(act)

	//发送交易
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

	//2.转账
	parmas2 := make(map[string]interface{})
	parmas2["from"] = Aks.DefaultPeerKS().Address()
	parmas2["to"] = "Address1"
	parmas2["value"] = 2500
	jbs2, _ := json.Marshal(parmas2)
	act = Act.NewPerfromAct(AyaChainDemoDapp_3, "transfer", string(jbs2))

	//签名
	tx = Aks.DefaultPeerKS().CreateSignedTx(act)

	//发送交易
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



	//3.查询余额
	act = Act.NewPerfromAct(AyaChainDemoDapp_3, "balanceOf", string(jbs))

	//签名
	tx = Aks.DefaultPeerKS().CreateSignedTx(act)

	//发送交易
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

	select {

	}
}

func main() {

	testDemo3()

	return

	nodeType := flag.String("t", "worker","")

	flag.Parse()

	peerType := DState.DappPeerType_Worker

	if *nodeType == "master" {
		peerType = DState.DappPeerType_Master
	}

	//生成一个Dapp状态机
	fristDemoState, err := DState.NewDappState(AyaChainDemoDapp_3)

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
				time.Sleep(time.Millisecond * 10)
				//time.Sleep(time.Second * 2)

				pmap := make(map[string]string)
				pmap["name"] = fmt.Sprintf("[VisiterNumber:%d]", txindex)
				pmapbs, _ := json.Marshal(pmap)

				var act Act.BaseActInf

				switch txindex % 3 {
				case 0 :
					act = Act.NewPerfromAct(AyaChainDemoDapp_3, "SayHello", string(pmapbs))

				case 1 :
					act = Act.NewPerfromAct(AyaChainDemoDapp_3, "GiveMeALTable", "")

				case 2 :
					act = Act.NewPerfromAct(AyaChainDemoDapp_3, "GiveMeANumber", "")
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