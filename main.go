package main

import (
	DState "./statepool/dappstate"
	Act "./statepool/tx/act"
	Tx  "./statepool/tx"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-ipfs-api"
	"log"
	"strconv"
	"time"
)

func main() {

	fristDemoState := DState.NewDappState("QmP7htLz57Gz6jiCVnWQEYeRxr3V7CzVjnkjtSLWYL8seQ","QmbjTJhV7G1tdSURQGg54MfFtFG89jrWM1EzAwBEUs1wgT")

	defer func() {
		if err := fristDemoState.DestoryMFSEnv(); err != nil {
			panic(err)
		}
	}()

	fristDemoState.DestoryMFSEnv()

	if err := fristDemoState.InitDappMFSEnv(); err != nil {
		panic(err)
	}

	if err := fristDemoState.Daemon(); err == nil {

		fmt.Println("DappState Daemoning.")

		go func() {

			//模拟发送交易
			//生成地址和私钥
			key, err := crypto.GenerateKey()
			if err !=nil {
				fmt.Println(err)
			}

			// 私钥:64个字符
			privateKey := hex.EncodeToString(key.D.Bytes())
			fmt.Println(privateKey)

			// 得到地址：42个字符
			address := crypto.PubkeyToAddress(key.PublicKey).Hex()
			fmt.Println(address)

			txindex := 0

			for {

				txindex++

				time.Sleep(time.Second * 2)

				act := Act.NewPerfromAct("QmP7htLz57Gz6jiCVnWQEYeRxr3V7CzVjnkjtSLWYL8seQ", "main", []string{"Parmas1", strconv.Itoa(txindex)})

				//签名
				tx := Tx.NewTx(address, act)

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
				topics := "AyaChainListner.QmP7htLz57Gz6jiCVnWQEYeRxr3V7CzVjnkjtSLWYL8seQ.Tx.Commit"

				if txhex, err := tx.EncodeToHex(); err != nil {
					panic(err)
				} else {
					err = shell.NewLocalShell().PubSubPublish(topics, txhex)
				}
			}

		}()

		select {}

	} else {

		panic(err)

	}
}