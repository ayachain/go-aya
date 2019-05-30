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
	"github.com/labstack/gommon/log"
	"io"
	"time"
)

type AAppPath string

const (
	AAppPath_Script 			AAppPath = "/Script"
	AAppPath_Resources 	AAppPath = "/Resource"
	AAppPath_Index 			AAppPath = "/Index"
	AAppPath_Evn 				AAppPath = "/Evn"
)

func (e AAppPath) ToString() string {
	switch (e) {
	case AAppPath_Script			: return "Script"
	case AAppPath_Resources		: return "Resource"
	case AAppPath_Index			: return "Index"
	case AAppPath_Evn				: return "Evn"
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
	recvChannel chan iface.PubSubMessage
	onListen bool

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

	l := lua.NewAVMState(aappns, pbnode, ind)
	if l == nil {
		return nil, errors.New("alvm create failed")
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
		recvChannel:make(chan iface.PubSubMessage, 128),
	}


	if err := l.MFS_Mkdir("/Data", false); err != nil {
		log.Print("MFS_Mkdir Failed")
	}

	if cid, err := l.MFS_Flush("/"); err != nil {
		log.Print(cid.String())
	}

	//if !ap.Listen() {
	//	ap.State = AAppStat_Stoped
	//	return nil, fmt.Errorf("AApp is loaded but startlisten faild")
	//}

	return ap, nil
}


func (a *aapp) Listen() bool {

	var sum int = 0
	a.onListen = true

	topics := a.Info.GetChannelTopics()

	for _, v := range topics {

		subscribe, err := a.api.PubSub().Subscribe(a.ctx, v)
		if err != nil {
			//可能会有某个信道启动失败，但不影响整个AAPP运行，只要有一个成功则可以正常工作
			log.Warnf("Topic %v subscribe failed error is %v", v, err.Error())
		}

		sum ++
		go func() {

			for a.onListen {

				if msg, err := subscribe.Next(a.ctx); err != nil {
					subscribe.Close()
				} else {
					a.recvChannel <- msg
				}

			}

		}()

	}

	if sum > 0 {
		a.State = AAppStat_Daemon
		return true
	} else {
		return false
	}

}

func (a *aapp) Shutdown() {

	a.onListen = false
	a.ctxCancel()
	close(a.recvChannel)

}