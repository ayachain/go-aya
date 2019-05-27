package miner

import (
	"github.com/ayachain/go-aya-alvm"
	Avmm "github.com/ayachain/go-aya/avm/miner/module"
	Atx "github.com/ayachain/go-aya/statepool/tx"
)

const (
	AvmState_IDLE = 0
	AvmState_Buzy = 1
	AvmState_Dead = 2
)

type Avm struct {

	l 					*lua.LState
	//每个虚拟机仅有一个正在工作到矿工,虚拟机与Dapp无关，在需要到时候将Dapp装载到虚拟机中，并且配置好对应矿工则可以开始计算结果
	miner				MinerInf
	State				int
}

func (vm *Avm) GetL() *lua.LState {
	return vm.l
}

func NewAvm( mfsbase string ) *Avm {

	avm := &Avm{l:lua.NewState(lua.Options{
		CallStackSize:1024,
		RegistrySize:4096,
		SkipOpenLibs:true,
		AAppns:mfsbase,
	}), State:AvmState_IDLE, miner:&MNCMiner{}}

	Avmm.InjectionAyaModules(avm.l)

	return avm
}

func (vm *Avm) AAppns() string {
	return vm.GetL().Options.AAppns
}

func (vm *Avm) StartSyncMining(pendingBlock *Atx.Block) (r string, err error) {

	vm.State = AvmState_Buzy

	defer func() { vm.State = AvmState_IDLE }()

	//将作业提交给矿工
	if r, err := vm.miner.MiningBlock(vm, pendingBlock); err != nil {
		return "", err
	} else {
		return r, nil
	}
}