package dappstate

import (
	TX "../tx"
	"context"
	"github.com/ipfs/go-ipfs-api"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	"io/ioutil"
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
	broadcasterMap  map[string]Broadcaster	`json:"-"`

}

func NewDappState(shash string, bdhash string) (dstate *DappState) {

	dstate = &DappState{
		IPNSHash:shash,
		LatestBDHash:bdhash,
	}

	dstate.Pool = TX.NewTxPool()
	dstate.sh = shell.NewLocalShell()
	dstate.listnerMap = make(map[string]Listener)
	dstate.broadcasterMap = make(map[string]Broadcaster)

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

	if err = dstate.initListner(); err != nil {
		return err
	}

	if err = dstate.initBroadcast(); err != nil {
		return err
	}

	//启动主节点打包
	dstate.Pool.StartGenBlockDaemon()

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

func (dstate *DappState) initListner() error {

	//接收交易提交
	if err := dstate.AddListner(CreateListener(DappListner_TxCommit, dstate)); err != nil {
		return err
	}

	if err := dstate.AddListner(CreateListener(DappListner_Broadcast, dstate)); err != nil {
		return err
	}

	return dstate.startingListening()
}
func (dstate *DappState) initBroadcast() error {

	//打开Block广播频道
	bbc := CreateBroadcaster(DappBroadcaster_Block, dstate)
	if err := dstate.AddBroadcaster(bbc); err != nil {
		return err
	} else {
		//设置广播信道
		dstate.Pool.BlockBroadcastChan = bbc.Channel()
	}

	return dstate.startingBroadcasting()
}

//装载，卸载Dapp虚拟目录
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

	//4.加载主节点列表 在根目录的 /_data/_master 文件中
	if rep, err := ish.Request("files/read").
		Arguments(basePath + "/_data/_master").
		Send(ctx); err != nil {

			return err

		} else {

			//正确获取返回值，判断返回值中是否存在文件内容
			if rep.Error != nil {

				return errors.New( rep.Error.Message )

			} else {

				//正确内容
				if bs,err := ioutil.ReadAll(rep.Output); err != nil {

					return nil

				} else {

					//此处先简单按照","分割读出，后续应该改为一个Json，主节点列表应该含有签名，确认是Dapp的拥有者授权的主节点
					dstate.mpids = strings.Split( strings.Trim(string(bs), "\n") , ",")

				}

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
