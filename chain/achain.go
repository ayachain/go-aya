package chain

import (
	"context"
	"errors"
	"fmt"
	AMsgBlock "github.com/ayachain/go-aya/chain/message/block"
	AMsgInfo "github.com/ayachain/go-aya/chain/message/chaininfo"
	AMsgTx "github.com/ayachain/go-aya/chain/message/transaction"
	ACore "github.com/ayachain/go-aya/consensus/core"
	ACImpl "github.com/ayachain/go-aya/consensus/impls"
	"github.com/ayachain/go-aya/vdb"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-ipfs/core"
	"github.com/libp2p/go-libp2p-pubsub"
	"time"
)

const BroadCastChanSize = 128


var(
	ErrUnSupportRawMessage			= errors.New("unsupport raw message")
	ErrCantLinkToChainExpected		= errors.New("not found chain in Aya")
)


type AyaChain interface {

	OpenChannel() error
	ChainIdentifier() string
	SendRawMessage( coder AvdbComm.RawDBCoder ) error
	Disslink()

}

type aChain struct {

	AyaChain

	/// Because the name of the channel needs to be computed through the first block,
	/// in aya, linking any chain requires a clear and clear understanding of the
	/// complete data of the handed-down block. Of course, you can use a CID to
	/// represent this data, and then after network loading, restore the data of this
	/// block and connect.
	BlockZero ABlock.GenBlock


	/// We manage different "Header", "Block", "Transaction", "Receipt", "Assets" for
	/// each chain. They are stored in a MFS directory. In order to better interact with
	/// these data, we have prepared a VDB Services for each chain. You can use this
	/// service interface to carry out the current chain, the current block of data.
	VDBServices vdb.CVFS


	/// Because in Aya, all data must be stored by the core technology of ipfs, and some
	/// module configurations may depend on IPFSNode, we save this node instance in our
	/// chain structure, ready for use.
	INode *core.IpfsNode


	/// Our Abstract notary structure, in fact, is to better express the operation process
	/// of the consensus mechanism. We are compared to a "person" serving the node, and
	/// this person is the same on all nodes. For details, please refer to the corresponding
	/// documents.
	notary ACore.Notary


	/// Recording the channel string used in this chain communication broadcasting, each
	/// chain should have one or more channels, which can be established according to
	/// different node types and different responsibilities. We advocate using Hash as the
	/// channel label and adding versions, so that if there are bugs in the future, we can
	/// directly change the channel name by changing the version number. To prevent the
	/// bifurcation from continuing to run. Of course, don't forget that if a node that has
	/// no privileges and takes up the channel to send data, it should be blacklisted.
	channelTopics string


	/// When you want to close the link of this example, use it. All threads will listen
	/// for the end and then enter the terminator.generally, it is called in Disslink in the
	/// AChain interface.
	channelCloseFunc context.CancelFunc


	/// Each AChain's chain object, not only needs to accept the messages on the chain, but
	/// also must have the ability to send transactions. Then "Chan" is responsible for receiving
	/// messages and caching some messages.
	///
	/// The number of cached messages is defined in "BroadCastChanSize". When the chain is ready,
	/// a thread will say that these messages are
	/// taken out and processed. Whether to send or reject is decided by consensus mechanism.
	broadcastChan chan []byte
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
		broadcastChan:make(chan[]byte, BroadCastChanSize),
	}, nil

}


func (chain *aChain) OpenChannel() error {

	sub, err := chain.INode.PubSub.Subscribe( chain.channelTopics, nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel( context.Background() )
	chain.channelCloseFunc = cancel
	chain.notary.StartWorking()

	// listen receive channel
	go func ( ctx context.Context, subs *pubsub.Subscription, notary ACore.Notary ) {

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

	// listen send channel
	go func( ctx context.Context, rc chan []byte, ind *core.IpfsNode, topic string ) {

		select {
		case rawmsg := <- rc : {

			if err := ind.PubSub.Publish( topic, rawmsg ); err != nil {
				fmt.Println(err)
			}

		}

		case <- ctx.Done() : {
			return
		}


		default: {
			time.Sleep( time.Microsecond  * 100 )
		}

		}

	}( ctx, chain.broadcastChan, chain.INode, chain.channelTopics )

	return nil
}


func (chain *aChain) ChainIdentifier() string {
	return chain.BlockZero.ChainID
}


func (chain *aChain) SendRawMessage( coder AvdbComm.RawDBCoder ) error {

	switch coder.Prefix() {
	case AMsgBlock.MessagePrefix: {
		return ErrUnSupportRawMessage
	}

	case AMsgTx.MessagePrefix: {
		return chain.notary.SendTransaction(coder)
	}

	case AMsgInfo.MessagePrefix: {
		return ErrUnSupportRawMessage
	}

	default:
		return ErrUnSupportRawMessage
	}

}