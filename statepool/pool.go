package statepool

import (
	"container/list"
	DState "github.com/ayachain/go-aya/statepool/dappstate"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	Act "github.com/ayachain/go-aya/statepool/tx/act"
	"github.com/pkg/errors"
)

type StatePool struct {

	dapps map[string]*DState.DappState

	txqueue *list.List
}

var DappStatePool = &StatePool{dapps:make(map[string]*DState.DappState),txqueue:list.New()}


func (sp *StatePool) GetTxStatus(appns string, txhash string) ( bindex uint64, tx *Atx.Tx, stat int, req *Atx.TxReceipt ) {

	ds, inPool := sp.dapps[appns]

	if !inPool {
		return 0,nil, Atx.TxState_NotFound, nil
	} else {
		return ds.Pool.SearchTxStatus(txhash)
	}

}

func (sp *StatePool) AddDappStatDaemon(appns string) error {

	_, exist := sp.dapps[appns]

	if exist {
		return errors.New("Dapp State is are already exist.")
	}

	ds, err := DState.NewDappState(appns)

	if err != nil {
		return err
	}

	if err := ds.Daemon(DState.DappPeerType_Master); err != nil {
		return err
	}

	sp.dapps[appns] = ds

	return nil

}

func (sp *StatePool) PublishTx(tx *Atx.Tx) error {

	act := Act.BaseAct{}

	if err := act.DecodeFromHex(tx.ActHex); err != nil {
		return err
	}

	_,exist := sp.dapps[act.DPath]

	if !exist {
		return errors.New("Dapp State is not exist.")
	}

	txHex, err := tx.EncodeToHex()

	if err != nil {
		return err
	}

	sp.dapps[act.DPath].GetBroadcastChannel(DState.PubsubChannel_Tx) <- txHex

	ret := <- sp.dapps[act.DPath].GetBroadcastChannel(DState.PubsubChannel_Tx)

	if ret != nil {
		return ret.(error)
	}

	return nil
}