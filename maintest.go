package main

import (
	"encoding/json"
	"flag"
	"fmt"
	AvmStn "github.com/ayachain/go-aya/avm"
	AGateway "github.com/ayachain/go-aya/gateway"
	Aks "github.com/ayachain/go-aya/keystore"
	DState "github.com/ayachain/go-aya/statepool/dappstate"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	"time"

	"github.com/ipfs/go-ipfs"
)

const (
	AyaChainDemoDapp_Token = "QmVUaqfbeW3qNmrZqAbNrAin5aikYzpVv6GRt82Un28pW8"
)


func testDemo3() {

	//if err := DSP.DappStatePool.AddDappStatDaemon(AyaChainDemoDapp_Token); err != nil {
	//	panic(err)
	//}

	//开始交易
	//1.申请代币
	tx1str := fmt.Sprintf(`{"address":"%s", "amount":5000}`, Aks.DefaultPeerKS().Address())
	act := Act.NewPerfromAct(AyaChainDemoDapp_Token, "giveMeSomeToken", tx1str)
	tx := Aks.DefaultPeerKS().CreateSignedTx(act)
	txstr, _:= json.Marshal(tx)
	fmt.Println("giveMeSomeToken: " + string(txstr))

	//2.转账
	tx2str := fmt.Sprintf(`{"from":"%s", "to":"0x88FFe3F7b26F0CEd6945BDA4e8621EC107049CE1", "value":100}`, Aks.DefaultPeerKS().Address())
	act = Act.NewPerfromAct(AyaChainDemoDapp_Token, "transfer", tx2str)
	tx = Aks.DefaultPeerKS().CreateSignedTx(act)
	txstr, _ = json.Marshal(tx)
	fmt.Println( "transfer: " + string(txstr))

	//3.查询
	tx3str := `{"from":"0x88FFe3F7b26F0CEd6945BDA4e8621EC107049CE1"}`
	act = Act.NewPerfromAct(AyaChainDemoDapp_Token, "transfer", tx3str)
	tx = Aks.DefaultPeerKS().CreateSignedTx(act)
	txstr, _ = json.Marshal(tx)
	fmt.Println( "balanceof: " + string(txstr))

	//4.info
	tx4str := `{"from":"0x88FFe3F7b26F0CEd6945BDA4e8621EC107049CE1"}`
	act = Act.NewPerfromAct(AyaChainDemoDapp_Token,"addressInfo", tx4str)
	tx = Aks.DefaultPeerKS().CreateSignedTx(act)
	txstr, _ = json.Marshal(tx)
	fmt.Println( "addressInfo: " + string(txstr))

}

func test2() {

	nodeType := flag.String("t", "worker","")

	flag.Parse()

	peerType := DState.DappPeerType_Worker

	if *nodeType == "master" {
		peerType = DState.DappPeerType_Master
	}

	//生成一个Dapp状态机
	fristDemoState, err := DState.NewDappState(AyaChainDemoDapp_Token)

	if err != nil {
		panic(err)
	}

	//启动状态机守护线程，当中有主题的监听
	if err := fristDemoState.Daemon(peerType); err == nil {

		//log.Println("AyaChain Daemon is ready.")

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
					act = Act.NewPerfromAct(AyaChainDemoDapp_Token, "SayHello", string(pmapbs))

				case 1 :
					act = Act.NewPerfromAct(AyaChainDemoDapp_Token, "GiveMeALTable", "")

				case 2 :
					act = Act.NewPerfromAct(AyaChainDemoDapp_Token, "GiveMeANumber", "")
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

		AvmStn.DaemonWorkstation()
		//阻塞主线程
		select {}

	} else {

		panic(err)

	}
}

func mainTest() {

	//VM := miner.NewAvm("Demo4")
	//
	////if err := VM.GetL().DoFile("/Users/apple/Desktop/AyaDapp/Avm/main.lua"); err != nil {
	////	log.Panicln(err.Error())
	////}
	//
	//
	//VM.GetL().DoString(`
	//
	//	--m_loadfile()
	//	--introduce()
	//
	//	--local Math = require "Operation"
	//	--print("Div:" .. Math.Div(21,7))
	//
	//`)
	//
	//return

	AvmStn.DaemonWorkstation()
	AGateway.DaemonHttpGateway()

	fmt.Printf(ipfs.ApiVersion)

	testDemo3()

	select {

	}
	return
}

//func main() {
//
//	if _, err := aapp.NewAApp("QmNxAb3GttYoUv98f6iCqnbqveApYX6dAHhxfdnDQi5Cqo"); err != nil {
//		log.Error(err)
//	}
//
//}