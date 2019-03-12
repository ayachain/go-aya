package MemoryPool

import (
	"AyaChain/ChainStruct"
	"encoding/hex"
	"encoding/json"
)

type MemoryPool struct {
	TransactionChain ChainStruct.Chain
}

var MainPool = &MemoryPool{}

func (mp* MemoryPool) PushTransaction(transaction Transaction) (err error) {

	c, e := json.Marshal(transaction)

	if e != nil {
		return e
	}

	if mp.TransactionChain.BlockNumber() == 0 {
		mp.TransactionChain.GenerateChain( hex.EncodeToString(c) )
	} else {
		mp.TransactionChain.GenerateNewBlock( hex.EncodeToString(c) )
	}

	return nil
}