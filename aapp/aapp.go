package aapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	lua "github.com/ayachain/go-aya-alvm"
	"github.com/ipfs/go-ipfs/core"
	dag "github.com/ipfs/go-merkledag"
	"github.com/ipfs/interface-go-ipfs-core"
	"io"
	"log"
	"time"
)

type AAppPath string

const (
	AAppPath_Script 	AAppPath = "/Script"
	AAppPath_Resources 	AAppPath = "/Resource"
	AAppPath_Index 		AAppPath = "/Index"
	AAppPath_Evn 		AAppPath = "/Evn"
)

func (e AAppPath) ToString() string {
	switch (e) {
	case AAppPath_Script		: return "/Script"
	case AAppPath_Resources		: return "/Resource"
	case AAppPath_Index			: return "/Index"
	case AAppPath_Evn			: return "/Evn"
	default: return "UNKNOWN"
	}
}

type aapp struct {

	//状态
	State AAppStat
	//基本信息
	Info info
	//创建时间
	CreateTime int64
	//虚拟机
	Avm *lua.LState

	ctx context.Context
	ctxCancel context.CancelFunc
	api iface.CoreAPI
	ind *core.IpfsNode
	onListen bool

	recvMsgChan chan iface.PubSubMessage
	shutdownChan chan bool
	broadcastChan chan []byte

}

func NewAApp( aappns string, api iface.CoreAPI, ind *core.IpfsNode ) ( ap *aapp, err error ) {
	
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if ap == nil {
			cancel()
		}
	}()

	path, err := api.Name().Resolve(ctx, aappns)
	if err != nil {
		return nil, fmt.Errorf("AAppns %v Resolve failed", aappns )
	}

	//Create MFS
	vroot, err := api.ResolveNode(ctx, path)
	if err != nil {
		return nil, err
	}

	pbnode, ok := vroot.(*dag.ProtoNode)
	if !ok {
		return nil, dag.ErrNotProtobuf
	}

	l, err := lua.NewAVMState(ctx, aappns, pbnode, ind)
	if err != nil {
		return nil, fmt.Errorf("Alvm create fialed : %v", err.Error())
	}

	fir, err := l.MFS_LookupFile("/Evn/info.json")
	if err != nil {
		return nil, err
	}

	bs, err := l.MFS_ReadAll(fir, io.SeekStart)
	if err != nil {
		return nil, err
	}

	var inf info
	if err := json.Unmarshal(bs, &inf); err != nil {
		return nil, errors.New(`"/Evn/info.json" Unmarshal failed`)
	}

	ap = &aapp{
		Avm:l,
		State:AAppStat_Loaded,
		CreateTime:time.Now().Unix(),
		Info:inf,
		ctx:ctx,
		ctxCancel:cancel,
		api:api,
		ind:ind,
	}

	if !ap.Daemon() {
		ap.State = AAppStat_Stoped
		return nil, fmt.Errorf("AApp is loaded but startlisten faild")
	}

	return ap, nil
}

func (a *aapp) Daemon() bool {

	ctx, cancel := context.WithCancel(context.Background())

	a.shutdownChan = make(chan bool)
	a.broadcastChan = make(chan []byte, 128)
	a.recvMsgChan = make(chan iface.PubSubMessage, 128)

	subscribe, err := a.api.PubSub().Subscribe( a.ctx, a.Info.GetChannelTopic() )

	if err != nil {
		return false
	}

	go func() {

		select {

		case bsmsg := <- a.broadcastChan:

			if err := a.api.PubSub().Publish(ctx, a.Info.GetChannelTopic(), bsmsg); err != nil {
				a.shutdownChan <- true
			}

		case <- a.shutdownChan:
			cancel()

		case <- ctx.Done():
			close(a.shutdownChan)
			close(a.broadcastChan)
			close(a.recvMsgChan)
			subscribe.Close()
			return

		case <- a.ctx.Done():
			cancel()

		case msg := <- a.recvMsgChan:
			//处理消息
			log.Print( string(msg.Data()) )
		}

	}()

	go func() {

		for {

			if msg, err := subscribe.Next(a.ctx); err != nil {
				log.Printf("AApp:%v subscribe shutdowned.", a.Info.AAppns)
				return
			} else {
				a.recvMsgChan <- msg
			}

		}

	}()

	return true
}

func (a *aapp) Shutdown() {
	a.shutdownChan <- true
}