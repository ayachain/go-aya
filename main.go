package main

import (
	"./ChainStruct"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-ipfs-api"
	"time"
)

func main() {

	peerType := flag.String("t","master","Peer Type")

	flag.Parse()

	txPool := &ChainStruct.TransactionChain{}

	switch *peerType {
	case "master":

		txPool.SetPeerType(ChainStruct.AyaPeerType_Master)
		if err := txPool.DeamonMasterPeer(); err != nil {
			panic(err)
			return
		}

		fmt.Printf("Master Peer Starting.")

	case "worker":

		txPool.SetPeerType(ChainStruct.AyaPeerType_Worker)
		if err := txPool.DeamonWorkerPeer(); err != nil {
			panic(err)
			return
		}

		fmt.Printf("Worker Peer Starting.")

	case "test":

		//发送交易
		go func() {

			key, err := crypto.GenerateKey()
			if err != nil {
				fmt.Println(err)
			}

			// 私钥:64个字符
			privateKey := hex.EncodeToString(key.D.Bytes())
			fmt.Println("PrivKey:" + privateKey)

			// 得到地址：42个字符
			address := crypto.PubkeyToAddress(key.PublicKey).Hex()
			fmt.Println("Address:" + address)

			for i := 0; i < 5000; i++ {

				time.Sleep(time.Duration(2) * time.Second)

				tx, err := ChainStruct.GenTransaction(
					address,
					string(i),
					ChainStruct.DappPerfrom{"path", "SayHello", []string{"Martin"}},
					key)

				if err != nil {
					panic(err)
					return
				}

				txHex, err := tx.EncodeToHex()

				if shell.NewLocalShell().PubSubPublish(ChainStruct.WTopics_AyaChainTransactionPool_Commit, txHex) != nil {
					continue
				}
			}

		}()

	}

	go func() {

		txPool.DumpPrint()
		time.Sleep(5000)

	}()

	select {

	}

	return

	////Block Test
	//chain := ChainStruct.GenerateChain("Content1")
	//chain.GenerateNewBlock("Content2")
	//chain.GenerateNewBlock("Content3")
	//chain.GenerateNewBlock("Content4")
	//chain.GenerateNewBlock("Content5")
	//
	//chain.DumpPrint()
}