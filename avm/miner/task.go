package miner

import (
	Atx "github.com/ayachain/go-aya/statepool/tx"
)

type MiningTask struct {
	DappNS			string
	PendingBlock 	*Atx.Block
	ResultChannel	chan interface{}
}

type TaskResult struct {
	DappNS			string
	PBlockHash		string
	RetBDHash		string
}

func (mt *MiningTask) CreateResult(rethash string) *TaskResult {
	return &TaskResult{mt.DappNS, mt.PendingBlock.GetHash(), rethash}
}