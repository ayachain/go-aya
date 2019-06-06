package chain

import (
	"context"
	"errors"
	"fmt"
	ACore "github.com/ayachain/go-aya/consensus/core"
	ACImpl "github.com/ayachain/go-aya/consensus/impls"
	"github.com/ayachain/go-aya/vdb"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-ipfs/core"
	"github.com/libp2p/go-libp2p-pubsub"
)

var(
	ErrCantLinkToChainExpected		= errors.New("not found chain in Aya")
)

type AyaChain interface {

	/// get a existed block by index
	GetBlockByIndex( i int ) (*ABlock.Block, error)

	/// get current confirmed block in this node
	GetCurBlock() (*ABlock.Block, error)

}

type aChain struct {
	AyaChain
	BlockZero ABlock.GenBlock
	VDBServices vdb.CVFS
	INode *core.IpfsNode

	notary ACore.Notary
	channelTopics string
	channelCloseFunc context.CancelFunc
}

func LinkChain( genBlock ABlock.GenBlock, ind *core.IpfsNode ) (AyaChain, error) {

	vdbfs, err := vdb.LinkVFS( genBlock.GetExtraDataCid(), ind )
	if err != nil {
		return nil, ErrCantLinkToChainExpected
	}

	topics := fmt.Sprintf("Aya 0.0.1_%v", genBlock.ChainID)
	topics = crypto.Keccak256Hash([]byte(topics)).String()

	// Create consensus norary
	norary, err := ACImpl.CreateNotary( genBlock.Consensus, vdbfs, ind )
	if err != nil {
		vdbfs.Close()
		return nil, err
	}

	return &aChain{
		BlockZero:genBlock,
		VDBServices:vdbfs,
		INode:ind,
		notary:norary,
	}, nil

}

func (chain *aChain) OpenChannel(  ) error {

	sub, err := chain.INode.PubSub.Subscribe( chain.channelTopics, nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel( context.Background() )
	chain.channelCloseFunc = cancel
	chain.notary.StartWorking()

	go func ( ctx context.Context,  subs *pubsub.Subscription, notary ACore.Notary ) {

		for {

			msg, err := subs.Next( ctx )
			if err != nil {
				fmt.Printf("Channel %v error for closed", err)
				return
			}

			if err := notary.OnReceiveMessage(msg); err != nil {
				fmt.Println(err)
			}

		}

	}( ctx, sub, chain.notary )

	return nil
}

func (chain *aChain) ChainIdentifier() string {
	return chain.BlockZero.ChainID
}