package chain

import (
	"context"
	"errors"
	"fmt"
	"github.com/ayachain/go-aya/chain/txpool"
	ACore "github.com/ayachain/go-aya/consensus/core"
	ACIMPL "github.com/ayachain/go-aya/consensus/impls"
	"github.com/ayachain/go-aya/vdb"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ayachain/go-aya/vdb/indexes"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/prometheus/common/log"
)

var(
	ErrAlreadyExistConnected				= errors.New("chan already exist connected")
	ErrCantLinkToChainExpected		= errors.New("not found chain in Aya")
)


type AyaChain interface {

	Disconnect()
	CVFServices() vdb.CVFS
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

func AddChainLink( ctx context.Context, genBlock *ABlock.GenBlock, ind *core.IpfsNode, acc EAccount.Account ) error {

	_, exist := chains[genBlock.ChainID]
	if exist {
		return ErrAlreadyExistConnected
	}

	var baseCid cid.Cid
	var err error

	idxs := indexes.CreateServices(ind, genBlock.ChainID)
	idx, err := idxs.GetLatest()
	_ = idxs.Close()

	if err == nil && idx.BlockIndex > 0 {

		baseCid = idx.FullCID

		fmt.Printf("LatestIndex:%06d CID:%v\n", idx.BlockIndex, idx.FullCID)

	} else {

		baseCid, err = vdb.CreateVFS(genBlock, ind)
		if err != nil {
			return err
		}
	}

	vdbfs, err := vdb.LinkVFS(genBlock.ChainID, baseCid, ind )
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

	ac := &aChain{
		INode:ind,
		Notary:notary,
		ChainId:genBlock.ChainID,
	}

	// config txpool
	ac.TxPool = txpool.NewTxPool( ind, genBlock, vdbfs, notary, acc)

	chains[genBlock.ChainID] = ac

	poolCtx, cancel := context.WithCancel(context.Background())

	ac.ctxCancel = cancel

	go func() {

		select {
		case <- ctx.Done():

			cancel()

			delete(chains, ac.ChainId)

		}

	}()

	if err := ac.TxPool.PowerOn(poolCtx); err != nil {
		return err
	}

	return nil
}

func GetChainByIdentifier(chainId string) AyaChain{
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

func (chain *aChain) PublishTx(tx *ATx.Transaction) error {
	return chain.TxPool.PublishTx( tx )
}

func (chain *aChain) CVFServices() vdb.CVFS {
	return chain.TxPool.ReadOnlyCVFS()
}