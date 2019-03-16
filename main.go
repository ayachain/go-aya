package main

import (
	Atx "github.com/ayachain/go-aya/statepool/tx"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	DState "github.com/ayachain/go-aya/statepool/dappstate"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-ipfs-api"
	"log"
	"strconv"
	"time"
)

func main() {

	//模拟发送交易
	//生成地址和私钥
	key, err := crypto.GenerateKey()
	if err !=nil {
		fmt.Println(err)
	}

	// 私钥:64个字符
	privateKey := hex.EncodeToString(key.D.Bytes())
	log.Println("PrivateKey : " + privateKey)

	// 得到地址：42个字符
	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	log.Println("Address : " + address)



	//生成一个Dapp状态机
	fristDemoState := DState.NewDappState("QmP7htLz57Gz6jiCVnWQEYeRxr3V7CzVjnkjtSLWYL8seQ","QmbjTJhV7G1tdSURQGg54MfFtFG89jrWM1EzAwBEUs1wgT")

	//在Main函数推出时执行，主要是清理虚拟目录
	defer func() {
		if err := fristDemoState.DestoryMFSEnv(); err != nil {
			panic(err)
		}
	}()

	//开始时执行，防止目录因为异常结束虚拟目录依然存在，导致InitDappMFSEnv失败
	fristDemoState.DestoryMFSEnv()

	//根据设置的SHash和DBHash生成虚拟目录环境
	if err := fristDemoState.InitDappMFSEnv(); err != nil {
		panic(err)
	}

	//启动状态机守护线程，当中有主题的监听，监听的Topics：AyaChainListner.QmP7htLz57Gz6jiCVnWQEYeRxr3V7CzVjnkjtSLWYL8seQ.Tx.Commit
	if err := fristDemoState.Daemon(); err == nil {

		log.Println("AyaChain Daemon is ready.")

		//启动测试线程发送交易
		go func() {

			txindex := 0

			for {

				txindex++

				//内部休眠时间为100毫秒 所以在保证在100毫秒内可以发送3比交易测试
				//time.Sleep(time.Millisecond * 35)
				time.Sleep(time.Second * 2)

				act := Act.NewPerfromAct("QmP7htLz57Gz6jiCVnWQEYeRxr3V7CzVjnkjtSLWYL8seQ", "main", []string{"Parmas1", strconv.Itoa(txindex)})

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

				//发送交易
				topics := "AyaChainChannel.QmP7htLz57Gz6jiCVnWQEYeRxr3V7CzVjnkjtSLWYL8seQ.Tx.Commit"

				if txhex, err := tx.EncodeToHex(); err != nil {
					panic(err)
				} else {
					err = shell.NewLocalShell().PubSubPublish(topics, txhex)
				}
			}

		}()

		//阻塞主线程
		select {}

	} else {

		panic(err)

	}
}