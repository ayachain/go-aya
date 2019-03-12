package main

import (
	"AyaChain/MemoryPool"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-ipfs-api"
)

func main() {

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

	for i := 0; i < 10; i++ {

		tx, err := MemoryPool.GenTransaction(
			address,
			string(i),
			MemoryPool.DappPerfrom{"path", "SayHello", []string{"Martin"}},
			key)

		if err != nil {
			panic(err)
			return
		}

		verify, err := tx.Verify()

		if err != nil {
			panic(err)
			return
		}

		if verify {
			err = MemoryPool.MainPool.PushTransaction(*tx)

			if err != nil {
				panic(err)
			}
		}
	}

	MemoryPool.MainPool.TransactionChain.DumpPrint()

	return

	////Block Test
	//chain := ChainStruct.GenerateChain("Content1")
	//chain.GenerateNewBlock("Content2")
	//chain.GenerateNewBlock("Content3")
	//chain.GenerateNewBlock("Content4")
	//chain.GenerateNewBlock("Content5")
	//
	//chain.DumpPrint()

	sh := shell.NewLocalShell()

	sub, err := sh.PubSubSubscribe("XMSP2P")

	if err != nil {
		panic(err)
	}

	go func() {
		for {

			m, e := sub.Next()

			if e != nil {
				panic(e)
			}

			//json,e := json.Marshal(m)

			if e != nil {
				panic(e)
			} else {
				fmt.Println(string(m.Data))
			}

		}
	}()

	fmt.Println("Subscribe XMSP2P Starting.")

	select {

	}

}