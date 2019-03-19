package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ayachain/go-aya/avm"
	DState "github.com/ayachain/go-aya/statepool/dappstate"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	"github.com/ethereum/go-ethereum/crypto"
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

	//模拟发送交易
	//生成地址和私钥
	key, err := crypto.GenerateKey()
	if err != nil {
		fmt.Println(err)
	}

	// 私钥:64个字符
	privateKey := hex.EncodeToString(key.D.Bytes())
	log.Println("PrivateKey : " + privateKey)

	// 得到地址：42个字符
	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	log.Println("Address : " + address)

	//生成一个Dapp状态机
	fristDemoState, err := DState.NewDappState("QmVUaqfbeW3qNmrZqAbNrAin5aikYzpVv6GRt82Un28pW8")

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

				act := Act.NewPerfromAct("QmVUaqfbeW3qNmrZqAbNrAin5aikYzpVv6GRt82Un28pW8", "main", []string{"Parmas1", strconv.Itoa(txindex)})

				//签名
				tx := Atx.NewTx(address, act)

				sig, err := crypto.Sign(crypto.Keccak256(tx.GetActHash()), key)

				if err != nil {
					panic(err)
				}

				tx.Signature = "0x" + hex.EncodeToString(sig)

				//验证
				if !tx.VerifySign() {
					log.Println("Verify Faild.")
					return
				}

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