package chain

import (
	"context"
	"errors"
	"github.com/ayachain/go-aya/chain/txpool"
	ACore "github.com/ayachain/go-aya/consensus/core"
	ACIMPL "github.com/ayachain/go-aya/consensus/impls"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ayachain/go-aya/vdb"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	ATransaction "github.com/ayachain/go-aya/vdb/transaction"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-ipfs/core"
)

const BroadCastChanSize = 128

var(
	ErrAlreadyExistConnected				= errors.New("chan already exist connected")
	ErrUnSupportRawMessage			= errors.New("unsupport raw message")
	ErrCantLinkToChainExpected		= errors.New("not found chain in Aya")
)


type AyaChain interface {

	Test() error
	OpenChannel() error
	ChainIdentifier() string
	Disslink()
	CVFServices() vdb.CVFS
	IndexOf( idx uint64 ) (*indexes.Index, error)
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

	baseCid, err := vdb.CreateVFS(genBlock, ind)
	if err != nil {
		return err
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
	ac.TxPool = txpool.NewTxPool( tpctx, ind, genBlock.ChainID, vdbfs, notary, acc)

	if err := ac.OpenChannel(); err != nil {
		return err
	}

	chains[genBlock.ChainID] = ac

	return nil
}

func GetChainByIdentifier(chainId string) AyaChain{
	return chains[chainId]
}

func (chain *aChain) OpenChannel() error {
	return chain.TxPool.PowerOn()
}

func (chain *aChain) SendRawMessage( coder AvdbComm.AMessageEncode ) error {
	return chain.TxPool.DoBroadcast( coder )
}

func (chain *aChain) IndexOf( idx uint64 ) (*indexes.Index, error) {

	index, err := chain.TxPool.ReadOnlyCVFS().Indexes().GetIndex(idx)

	if err != nil {
		return nil, err
	}

	return index, nil
}

func (chain *aChain) CVFServices() vdb.CVFS {
	return chain.TxPool.ReadOnlyCVFS()
}

var tid uint64 = 0
func (chain *aChain) Test() error {

	tx := &ATransaction.Transaction{}
	tx.BlockIndex = 0
	tx.From = common.HexToAddress("0xfC8Bc1E33131Bd9586C8fB8d9E96955Eb1210C67")
	tx.To = common.HexToAddress("0x341f244DDd50f51187a6036b3BDB4FCA9cAFeE16")
	tx.Value = 10000
	tx.Data = nil
	tx.Steps = 50
	tx.Price = 1
	tx.Tid = tid

	tid ++

	acc, err := AKeyStore.FindAccount("0xfC8Bc1E33131Bd9586C8fB8d9E96955Eb1210C67")
	if err != nil {
		return err
	}

	if err := AKeyStore.SignTransaction(tx, acc); err != nil {
		return err
	}

	return chain.TxPool.DoBroadcast(tx)
	//cbs := tx.RawMessageEncode()
	//
	//if signmsg, err := AKeyStore.CreateMsg(cbs, acc); err != nil {
	//	return err
	//} else {
	//
	//	err := chain.TxPool.RawMessageSwitch(signmsg)
	//	if err != nil {
	//		return err
	//	}
	//
	//}
	//
	//return nil
}