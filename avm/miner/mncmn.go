package miner

import (
	Atx "github.com/ayachain/go-aya/statepool/tx"
	Autils "github.com/ayachain/go-aya/utils"
)

//Master Node Consensus Miner
//主节点共识
//共识机制说明：
//所有的交易可以通过任意相同Dapp的节点进行提交，提交到主节点后，主节点负责计算结果然后分发到全网
//优点：共识可靠，相对并发速度仅与主节点到计算能力和网络延迟有关系，不会存在拜占庭节点，因为其他到节点仅备份数据
type MNCMiner struct {
	MinerInf
}

func (m* MNCMiner) MiningBlock(vm *Avm, b* Atx.Block) (r string, err error) {

	pblk, err := Atx.ReadBlock(b.Prev)

	if err != nil {
		return "", err
	}

	Autils.AFMS_ReloadDapp(pblk.BDHash, pblk.BDHash)

	defer func() {

		Autils.AFMS_RemovePath(pblk.BDHash)

	}()

	codestr, err := Autils.AFMS_ReadDappCode(pblk.BDHash)

	if err != nil {
		return "", err
	}

	if err := vm.l.DoString(codestr); err != nil {
		return "", err
	}

	dirstat, err := Autils.AFMS_PathStat(pblk.BDHash)

	if err != nil {
		return dirstat.Hash,err
	} else {
		return "",err
	}

}