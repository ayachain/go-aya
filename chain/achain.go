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
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"sync"
)

const BroadCastChanSize = 128

var(
	ErrAlreadyExistConnected				= errors.New("chan already exist connected")
	ErrUnSupportRawMessage			= errors.New("unsupport raw message")
	ErrCantLinkToChainExpected		= errors.New("not found chain in Aya")
)


type AyaChain interface {

	Connect() error
	Disconnect()
	CVFServices() vdb.CVFS
	SendRawMessage( coder AvdbComm.AMessageEncode ) error

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
	ctx context.Context
	ctxCancel context.CancelFunc

	TxPool* txpool.ATxPool
}

var chains = make(map[string]AyaChain)

func AddChainLink( genBlock *ABlock.GenBlock, ind *core.IpfsNode, acc EAccount.Account ) error {

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

		fmt.Printf("Readed LatestIndex:%v CID:%v", idx.BlockIndex, idx.FullCID)

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

	notary, err := ACIMPL.CreateNotary( genBlock.Consensus, ind )
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	ac := &aChain{
		INode:ind,
		Notary:notary,
		ctx:ctx,
		ctxCancel:cancel,
	}

	// config txpool
	tpctx, _ := context.WithCancel(ctx)
	ac.TxPool = txpool.NewTxPool( tpctx, ind, genBlock, vdbfs, notary, acc)

	if err := ac.Connect(); err != nil {
		return err
	}

	chains[genBlock.ChainID] = ac

	return nil
}

func GetChainByIdentifier(chainId string) AyaChain{
	return chains[chainId]
}

func DisconnectionAll() <- chan bool {

	var rchan chan bool

	go func() {

		fmt.Println("Waiting Disconnection")

		wg := sync.WaitGroup{}
		wg.Add(len(chains))

		for _, chain := range chains {

			chain.Disconnect()
			wg.Done()
		}

		wg.Wait()

		fmt.Println("Waiting Disconnection Done")

		rchan <- true

	}()

	return rchan
}

func (chain *aChain) Connect() error {

	return chain.TxPool.PowerOn(chain.ctx)
}

func (chain *aChain) Disconnect() {

	chain.ctxCancel()

	<- chain.ctx.Done()

	chain.TxPool.Close()

	return
}

func (chain *aChain) SendRawMessage( coder AvdbComm.AMessageEncode ) error {
	return chain.TxPool.DoBroadcast( coder )
}

func (chain *aChain) CVFServices() vdb.CVFS {
	return chain.TxPool.ReadOnlyCVFS()
}