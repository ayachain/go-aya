package miner

import (
	"encoding/json"
	"fmt"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	Autils "github.com/ayachain/go-aya/utils"
	"github.com/pkg/errors"
	"github.com/yuin/gopher-lua"
	LJson "layeh.com/gopher-json"
	"log"
	"time"
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

	stime := time.Now().Unix()

	if err != nil {
		return "", err
	}

	//1.载入当前块app的所有数据,默认的flush=false
	Autils.AFMS_ReloadDapp(pblk.BDHash, vm.DappNS)

	//1.写入检索, 检索文件位于 对应块数据下的 /_index/_bindex，使用IPFSHash作为间隔直接写入,读取检索使用offset和hashsize
	if err := m.writingBlockIndex(vm.DappNS, b); err != nil {
		return "", err
	}

	codestr, err := Autils.AFMS_ReadDappCode(vm.DappNS)

	if err != nil {
		return "", err
	}

	if err := vm.l.DoString(codestr); err != nil {

		return "", err

	} else {

		//虚拟机载入完毕，开始计算交易,必须是单线程，而且严格按照顺序执行，否则不通的顺序，不同的节点计算的结果会不一致导致无法出块
		//此处不再次验证签名，因为在主节点广播之前，已经完成了交易来源的签名验证，若工作节点在此强行修改参数，会直接导致计算结果不一致，从而被其他节点丢弃结果
		for i, tx := range b.Txs {

			pact := Act.PerfromAct{}

			if err := pact.DecodeFromHex(tx.ActHex); err != nil {

				if m.writeTxReceipt(vm.DappNS, b, i, "Error") != nil {
					//若发现无法写入，则发生了未知错误，此时没有任何矿工可以正常工作，应当直接放弃为此块计算最终结果
					return "", errors.New("MNCMiner : Can't write tx receipt content to mfs.")
				}

			} else {

				method := pact.Method
				parmas := pact.Parmas

				var ltbv lua.LValue

				if len(parmas) != 0 {

					ltbv, err = LJson.Decode(vm.l, []byte(parmas))

					if err != nil {

						if m.writeTxReceipt(vm.DappNS, b, i, "Parmas Parser Expection.") != nil {
							//若发现无法写入，则发生了未知错误，此时没有任何矿工可以正常工作，应当直接放弃为此块计算最终结果
							return "", errors.New("MNCMiner : Can't write tx receipt content to mfs.")
						} else {
							continue
						}

					}
				}

				if ltbv != lua.LNil {

					err = vm.l.CallByParam(lua.P {
						Fn:      vm.l.GetGlobal(method),
						NRet:    1,
						Protect: true,
					}, ltbv)

				} else {

					err = vm.l.CallByParam(lua.P {
						Fn:      vm.l.GetGlobal(method),
						NRet:    1,
						Protect: true,
					})
				}

				if err != nil {

					if m.writeTxReceipt(vm.DappNS, b, i, err.Error()) != nil {
						//若发现无法写入，则发生了未知错误，此时没有任何矿工可以正常工作，应当直接放弃为此块计算最终结果
						return "", err
					} else {
						continue
					}

				} else {

					if err := m.writeTxReceipt(vm.DappNS, b, i, vm.l.Get(-1)); err != nil {
						return "", err
					} else {
						vm.l.Pop(1)
						continue
					}

				}

			}

		}

	}

	if err := Autils.AFMS_FlushPath(vm.DappNS); err != nil {
		return "", errors.New("MNCMiner : Autils.AFMS_FlushPath Faild.")
	}

	dirstat, err := Autils.AFMS_PathStat(vm.DappNS)

	log.Printf("BlockIndex:%d Txs:%d Time:%d", b.Index, len(b.Txs), time.Now().Unix() - stime)

	if err != nil {
		return "",err
	} else {
		return dirstat.Hash, nil
	}

}

func (m* MNCMiner) writingBlockIndex(dappns string, b* Atx.Block) error {

	if b.Index <= 1 {
		//在第一块时直接创建块检索文件
		return Autils.AFMS_CreateFile(dappns + "/_index", "_bindex", []byte(b.GetHash()))
	} else {
		//后续的则直接追写块的Hash值
		return Autils.AFMS_FileAppend(dappns + "/_index", "_bindex", []byte(b.GetHash()))
	}
}

//写入交易对应的返回结果
func (m* MNCMiner) writeTxReceipt( dappns string, b* Atx.Block, txindex int, data interface{} ) error {

	rep := &Atx.TxReceipt{}
	rep.BlockIndex = b.Index
	rep.TxHash = b.Txs[txindex].GetSha256Hash()
	rep.Tx = b.Txs[txindex]
	rep.Status = "Confirm"

	t, islvalue := data.(lua.LValue)

	if islvalue {
		switch t.Type() {
		case lua.LTBool:
			rep.Response = lua.LVAsBool(t)

		case lua.LTString:
			rep.Response = lua.LVAsString(t)

		case lua.LTNumber:
			rep.Response = lua.LVAsNumber(t)

		case lua.LTTable:

			if bs,err := LJson.Encode(t); err != nil {

				tbj := make(map[string]string)

				if err := json.Unmarshal(bs, tbj); err == nil {
					rep.Response = tbj
				}
			}

		default :
			rep.Response = nil
		}

	} else {
		rep.Response = data
	}

	bs,err := rep.MarshalJson()

	if err != nil {
		return err
	}

	if err := Autils.AFMS_CreateFile(dappns, "/_receipt/" + rep.TxHash, bs); err != nil {
		return err
	}

	fmt.Println(string(bs))

	return nil
}