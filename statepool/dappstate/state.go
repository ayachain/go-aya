package dappstate

import (
	TX "../tx"
	"github.com/ipfs/go-ipfs-api"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	"strings"
)

//Dapp 状态机
type DappState struct {

	SourceHash		string
	LatestBDHash 	string
	Pool*			TX.TxPool
	sh*				shell.Shell				`json:"-"`
	//Master Peer IDS 可信任的主节点ID
	mpids[]			string					`json:"-"`
	listnerMap 		map[string]Listener 	`json:"-"`
}

func NewDappState(shash string, bdhash string) (dstate *DappState) {

	dstate = &DappState{
		SourceHash:shash,
		LatestBDHash:bdhash,
	}

	dstate.Pool = TX.NewTxPool()
	dstate.sh = shell.NewLocalShell()

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
	if err = dstate.AddListner(CreateListener(DappListner_Broadcast, dstate)); err != nil {
		return err
	}

	return dstate.StartingListening()
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