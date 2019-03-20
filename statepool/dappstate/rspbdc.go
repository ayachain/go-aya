package dappstate

import (
	"github.com/ayachain/go-aya/avm/miner"
	Aks "github.com/ayachain/go-aya/keystore"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	"github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
)

type RspBroadCaseter struct {
	BaseBroadcaster
}

func NewRspBroadCaseter(ds* DappState) Broadcaster {

	bbc := &RspBroadCaseter{}
	bbc.state = ds
	bbc.typeCode = PubsubChannel_Rsp
	bbc.topics = BroadcasterTopicPrefix + ds.DappNS + ".Block.BDHashReply"
	bbc.channel = make(chan interface{})
	bbc.handleDelegate = bbc.bcHandleDelegate

	return bbc
}

func (bb *RspBroadCaseter) bcHandleDelegate(v interface{}) error {

	ret, ok := v.(*miner.TaskResult)

	if !ok {
		return errors.New("RspBroadCaseter : Channel recv object is not kind of miner.TaskResult")
	}

	actHex,err := Act.NewTxRspAct(ret.DappNS, ret.PBlockHash, ret.RetBDHash).EncodeToHex()

	if err != nil {
		return err
	}

	mtx := Atx.Tx{ActHex:actHex}

	if err := Aks.DefaultPeerKS().SignTx(&mtx); err != nil {
		return err
	}

	mtxhex, err := mtx.EncodeToHex()

	if err != nil {
		return err
	} else {
		return shell.NewLocalShell().PubSubPublish(bb.GetTopics(), mtxhex)
	}

}