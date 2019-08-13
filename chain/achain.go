package chain

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ayachain/go-aya/chain/txpool"
	ACore "github.com/ayachain/go-aya/consensus/core"
	ACIMPL "github.com/ayachain/go-aya/consensus/impls"
	"github.com/ayachain/go-aya/vdb"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ayachain/go-aya/vdb/indexes"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/core"
	"github.com/prometheus/common/log"
)

var(
	ErrAlreadyExistConnected		= errors.New("chan already exist connected")
	ErrCantLinkToChainExpected		= errors.New("not found chain in Aya")
)

const (
	AChainMapKey = "aya.chain.list.map"
)

type AyaChain interface {

	Disconnect()

	CVFServices() vdb.CVFS

	GetTxPool() *txpool.ATxPool

	PublishTx(tx *ATx.Transaction) error
}


type aChain struct {

	AyaChain

	/// Because in Aya, all data must be stored by the core technology of ipfs, and some
	/// module configurations may depend on IPFSNode, we save this node instance in our
	/// chain structure, ready for use.
	INode *core.IpfsNode

	/// Our Abstract notary structure, in fact, is to better express the operation process
	/// of the consensus mechanism. We are compared to a "person" serving the node, and
	/// this person is the same on all nodes. For details, please refer to the corresponding
	/// documents.
	Notary ACore.Notary

	/// When you want to close the link of this example, use it. All threads will listen
	/// for the end and then enter the terminator.generally, it is called in Disslink in the
	/// AChain interface.
	ctxCancel context.CancelFunc

	TxPool* txpool.ATxPool

	ChainId string
}

var chains = make(map[string]AyaChain)

func Conn( ctx context.Context, chainId string, ind *core.IpfsNode, acc EAccount.Account ) error {

	maplist, err := GenBlocksMap(ind)
	if err != nil {
		return err
	}

	genBlock, exist := maplist[chainId]
	if !exist {
		return errors.New(`can't find the corresponding chain, did you not execute "add"`)
	}

	_, exist = chains[genBlock.ChainID]
	if exist {
		return ErrAlreadyExistConnected
	}

	idxs := indexes.CreateServices(ind, genBlock.ChainID, false)
	if idxs == nil {
		return errors.New(`can't find the corresponding chain, did you not execute "add"`)
	}
	defer func() {
		if err := idxs.Close(); err != nil {
			log.Error(err)
		}
	}()

	lidx, err := idxs.GetLatest()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Infof("Read Block: %08d CID: %v", lidx.BlockIndex, lidx.FullCID.String())

	vdbfs, err := vdb.LinkVFS(genBlock.ChainID, ind, idxs)
	if err != nil {
		return ErrCantLinkToChainExpected
	}
	defer func() {
		if err := vdbfs.Close(); err != nil {
			log.Error(err)
		}
	}()

	notary, err := ACIMPL.CreateNotary( genBlock.Consensus, ind )
	if err != nil {
		return err
	}

	poolCtx, cancel := context.WithCancel( ctx )

	ac := &aChain{
		INode:ind,
		Notary:notary,
		ChainId:genBlock.ChainID,
		TxPool:txpool.NewTxPool( ind, genBlock, vdbfs, notary, acc),
		ctxCancel:cancel,
	}

	chains[genBlock.ChainID] = ac

	go func() {

		select {
		case <- ctx.Done():
			delete(chains, ac.ChainId)

		case <- poolCtx.Done():
			delete(chains, ac.ChainId)

		}

	}()

	if err := ac.TxPool.PowerOnAndLoop(poolCtx); err != nil {
		return err
	}

	return nil
}

func GenBlocksMap( ind *core.IpfsNode ) (map[string]*ABlock.GenBlock, error) {

	dsk := datastore.NewKey(AChainMapKey)
	val, err := ind.Repo.Datastore().Get(dsk)

	if err != nil {
		if err != datastore.ErrNotFound {
			return nil, err
		}
	}

	rmap := make(map[string]*ABlock.GenBlock)
	if val != nil {

		if err := json.Unmarshal( val, &rmap ); err != nil {
			return nil, err
		}

	}

	return rmap, nil
}

func AddChain( genBlock *ABlock.GenBlock, ind *core.IpfsNode, r bool ) error {

	maplist, err := GenBlocksMap(ind)
	if err != nil {
		return err
	}

	_, exist := maplist[genBlock.ChainID]
	if exist && !r {
		return errors.New("chain are already exist")
	}

	// Create indexes
	idxServer := indexes.CreateServices(ind, genBlock.ChainID, r)
	if idxServer == nil {
		return errors.New("create chain indexes services failed")
	}
	defer func() {
		if err := idxServer.Close(); err != nil {
			log.Error(err)
		}
	}()

	// Create CVFS and write genBlock
	if _, err := vdb.CreateVFS(genBlock, ind, idxServer); err != nil {
		return err
	}

	maplist[genBlock.ChainID] = genBlock
	bs, err := json.Marshal(maplist)
	if err != nil {
		return err
	}

	dsk := datastore.NewKey(AChainMapKey)
	if err := ind.Repo.Datastore().Put( dsk, bs ); err != nil {
		return err
	}

	return nil
}

func GetChainByIdentifier(chainId string) AyaChain {
	return chains[chainId]
}

func DisconnectionAll() {

	for _, chain := range chains {
		chain.Disconnect()
	}

}

func (chain *aChain) Disconnect() {
	chain.ctxCancel()
}

func (chain *aChain) GetTxPool() *txpool.ATxPool {
	return chain.TxPool
}

func (chain *aChain) PublishTx(tx *ATx.Transaction) error {
	return chain.TxPool.PublishTx( tx )
}

func (chain *aChain) CVFServices() vdb.CVFS {
	return chain.TxPool.ReadOnlyCVFS()
}