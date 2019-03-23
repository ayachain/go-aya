package dappstate

import (
	"encoding/json"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	"github.com/ipfs/go-ipfs-api"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	"strings"
)

const (
	DappPeerType_Master = 0
	DappPeerType_Worker = 1
)

//Dapp 状态机
type DappState struct {

	DappNS			string
	LatestBDHash 	string
	Pool*			Atx.TxPool
	//Master Peer IDS 可信任的主节点ID
	mpids[]			string					`json:"-"`
	listnerMap 		map[string]Listener 	`json:"-"`
	broadcasterMap  map[string]Broadcaster	`json:"-"`

}

func NewDappState(dappns string) (dstate *DappState, err error) {

	aappbs, err := shell.NewLocalShell().BlockGet(dappns)

	if err != nil {
		return nil, err
	}

	app := &Aapp{}

	if json.Unmarshal(aappbs, app) != nil {
		return nil, err
	}

	dstate = &DappState{
		DappNS:dappns,
		LatestBDHash:app.BDHash,
		mpids:app.MasterNode,
	}

	dstate.Pool = Atx.NewTxPool(dappns, Atx.NewBlock(0,"",nil,app.BDHash))

	dstate.listnerMap = make(map[string]Listener)
	dstate.broadcasterMap = make(map[string]Broadcaster)

	return dstate, nil
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
func (dstate *DappState) Daemon(peerType int) error {

	var err error

	defer func() {

		if err != nil {
			dstate.Clean()
		}

	}()

	if err = dstate.initListner(peerType); err != nil {
		return err
	}

	if err = dstate.initBroadcast(peerType); err != nil {
		return err
	}

	//启动主节点打包
	if peerType == DappPeerType_Master {
		dstate.Pool.StartGenBlockDaemon()
	}

	return nil
}

//根据工厂中定义的枚举，获取广播的信道
func (dstate *DappState) GetBroadcastChannel(btype int) chan interface{} {

	for _, v := range dstate.broadcasterMap {

		if v.TypeCode() == btype {
			return v.Channel()
		}
	}

	return nil
}

func (dstate *DappState) AddListner(l Listener) error {

	if _,exist := dstate.listnerMap[l.GetTopics()]; exist {
		return errors.New(l.GetTopics() + " Listner instance are already exist.")
	}

	dstate.listnerMap[l.GetTopics()] = l

	return nil
}
func (dstate *DappState) initListner(peerType int) error {

	//Tx
	if err := dstate.AddListner(CreateListener(PubsubChannel_Tx, dstate)); err != nil {
		return err
	}

	//Block
	if err := dstate.AddListner(CreateListener(PubsubChannel_Block, dstate)); err != nil {
		return err
	}


	if peerType == DappPeerType_Master {

		//Rsp
		lsn := CreateListener(PubsubChannel_Rsp, dstate)

		if err := dstate.AddListner(lsn); err != nil {
			return err
		} else {
			lsn.(*RspListener).RspActOutChan = dstate.Pool.BlockBDHashChan
		}

	}

	return dstate.startingListening()
}
func (dstate *DappState) shutdownListening() {

	for _,v := range dstate.listnerMap {
		v.Shutdown()
	}

}
func (dstate *DappState) startingListening() error {

	var err error

	defer func() {

		if err != nil {
			dstate.shutdownListening()
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

func (dstate *DappState) AddBroadcaster(b Broadcaster) error {

	if _, exist := dstate.broadcasterMap[b.GetTopics()]; exist {
		return errors.New(b.GetTopics() + " Broadcaster instance are already exist.")
	}

	dstate.broadcasterMap[b.GetTopics()] = b

	return nil
}
func (dstate *DappState) initBroadcast(peerType int) error {

	//Tx
	tbc := CreateBroadcaster(PubsubChannel_Tx, dstate)
	if err := dstate.AddBroadcaster(tbc); err != nil {
		return err
	}

	//Rsp
	rspbc := CreateBroadcaster(PubsubChannel_Rsp, dstate)
	if err := dstate.AddBroadcaster(rspbc); err != nil {
		return err
	}

	//Block
	if peerType == DappPeerType_Master {
		//广播Block为主节点专属信道
		bbc := CreateBroadcaster(PubsubChannel_Block, dstate)
		if err := dstate.AddBroadcaster(bbc); err != nil {
			return err
		} else {
			//设置广播信道
			dstate.Pool.BlockBroadcastChan = bbc.Channel()
		}
	}

	return dstate.startingBroadcasting()
}
func (dstate *DappState) startingBroadcasting() error {

	var err error

	defer func() {

		if err != nil {
			dstate.shutdownBroadcasting()
		}

	}()

	for _, v := range dstate.broadcasterMap {

		if err = v.OpenChannel(); err != nil {
			return err
		}
	}

	return err

}
func (dstate *DappState) shutdownBroadcasting() {
	for _,v := range dstate.broadcasterMap {
		v.CloseChannel()
	}
}

func (dstate *DappState) Clean() {

	dstate.shutdownListening()
	dstate.listnerMap = map[string]Listener{}

	dstate.shutdownBroadcasting()
	dstate.broadcasterMap = map[string]Broadcaster{}
}
