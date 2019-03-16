package dappstate

import (
	TX "../tx"
	"context"
	"github.com/ipfs/go-ipfs-api"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	"strings"
)

//Dapp 状态机
type DappState struct {

	IPNSHash		string
	LatestBDHash 	string
	Pool*			TX.TxPool
	sh*				shell.Shell				`json:"-"`
	//Master Peer IDS 可信任的主节点ID
	mpids[]			string					`json:"-"`
	listnerMap 		map[string]Listener 	`json:"-"`
}

func NewDappState(shash string, bdhash string) (dstate *DappState) {

	dstate = &DappState{
		IPNSHash:shash,
		LatestBDHash:bdhash,
	}

	dstate.Pool = TX.NewTxPool()
	dstate.sh = shell.NewLocalShell()
	dstate.listnerMap = make(map[string]Listener)

	return dstate
}

func (dstate *DappState) ContainMaterPeer(id peer.ID) bool {

	for _,v := range dstate.mpids {

		if strings.EqualFold(v, id.String()) {
			return true
		}
	}

	return false
}

/*
启动对应Dapp的服务，服务内容包括
1.接受交易，接收到交易后向主节点提交交易，发送Tx
2.接受主节点广播的交易,接收Block，数据都为IPFSHASH
*/
func (dstate *DappState) Daemon() error {

	var err error

	defer func() {

		if err != nil {
			dstate.Clean()
		}

	}()

	//接收主节点广播的新块
	//if err = dstate.AddListner(CreateListener(DappListner_Broadcast, dstate)); err != nil {
	//	return err
	//}

	//接收交易提交
	if err = dstate.AddListner(CreateListener(DappListner_TxCommit, dstate)); err != nil {
		return err
	}

	//开始接收TxPool的Block广播信道
	go func() {

		for {

			boardcastBlockHash := <- dstate.Pool.BlockBoardcastChan



		}

	}()

	return dstate.StartingListening()
}

func (dstate *DappState) DestoryMFSEnv() error {

	ish := shell.NewLocalShell()

	basePath := "/" + dstate.IPNSHash

	if err := ish.Request("files/rm").
		Arguments(basePath).
		Option("r",true).
		Exec(context.Background(), nil); err != nil {
		return err
	}

	return nil
}

//装载App
func (dstate *DappState) InitDappMFSEnv() error {

	ish := shell.NewLocalShell()

	ctx := context.Background()

	basePath := "/" + dstate.IPNSHash

	//1.创建IPFS MFS 文件系统,文件名为dappPath应该为一个IPNS,使用IPFS作为文件夹名称
	if err := ish.Request("files/mkdir").
		Arguments(basePath).
		Exec(ctx, nil); err != nil {
		return err
	}

	//2.下载目录_dapp
	if err := ish.Request("files/cp").
		Arguments("/ipfs/" + dstate.IPNSHash, basePath + "/_dapp").
		Exec(ctx, nil); err != nil {
			return nil
	}

	//3.下载数据目录
	if err := ish.Request("files/cp").
		Arguments("/ipfs/" + dstate.LatestBDHash, basePath + "/_data").
		Exec(ctx, nil); err != nil {
			return err
	}

	return nil
}

func (dstate *DappState) Clean() {

	dstate.ShutdownListening()

	dstate.listnerMap = map[string]Listener{}
}

func (dstate *DappState) ShutdownListening() {

	for _,v := range dstate.listnerMap {
		v.Shutdown()
	}

}

//启动所有主题关注者
func (dstate *DappState) StartingListening() error {

	var err error

	defer func() {

		if err != nil {
			dstate.ShutdownListening()
		}

	}()

	for _, v := range dstate.listnerMap {

		if v.ThreadState() == ListennerThread_Stop {

			if err = v.StartListening(); err != nil {
				return err
			}
		}

	}

	return err
}

func (dstate *DappState) AddListner(l Listener) error {

	if _,exist := dstate.listnerMap[l.GetTopics()]; exist {
		return errors.New(l.GetTopics() + " Listner instance are already exist.")
	}

	dstate.listnerMap[l.GetTopics()] = l

	return nil
}