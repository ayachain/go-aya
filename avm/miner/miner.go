package miner

import (
	Atx "github.com/ayachain/go-aya/statepool/tx"
)

type MinerInf interface {
	MiningBlock(vm *Avm, b* Atx.Block) (r string, err error)
}