package msgcenter

import (
	"context"
	"errors"
	ASD "github.com/ayachain/go-aya/chain/sdaemon/common"
	"github.com/ayachain/go-aya/vdb"
	ACInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	"github.com/ayachain/go-aya/vdb/im"
	AMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMined "github.com/ayachain/go-aya/vdb/minined"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-ipfs/core"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/prometheus/common/log"
	"sync"
)

var (
	ErrPublishedContentEmpty = errors.New("published content is empty")
)

type MessageCenter interface {

	Push( msg *pubsub.Message, prefix string )

	PublishMessage( pmsg proto.Message, topic string ) error

	TrustMessage() <- chan interface{}

	Refresh()

	PowerOn( ctx context.Context, chainID string, ind *core.IpfsNode )
}


type aMessageCenter struct {

	MessageCenter

	cvfs vdb.CVFS

	cnf TrustedConfig

	/// Hash -> ReviewMessage
	reviewMsgMap sync.Map

	threshold uint64

	replay chan interface{}

	rmu sync.Mutex

	totalCount uint64

	ind *core.IpfsNode

	asd ASD.StatDaemon
}


func NewCenter( ind *core.IpfsNode, cvfs vdb.CVFS, cnf TrustedConfig, asd ASD.StatDaemon) MessageCenter {

	c := &aMessageCenter{
		cvfs:cvfs,
		cnf:cnf,
		replay:make(chan interface{}),
		asd:asd,
		ind:ind,
	}

	c.Refresh()

	return c
}

func (mc *aMessageCenter) Refresh() {

	mc.rmu.Lock()
	defer mc.rmu.Unlock()

	p := uint64( mc.cnf.VotePercentage * 100 )

	mc.threshold = uint64( mc.cvfs.Nodes().GetSuperMaterTotalVotes() / 100 * p )
}

func (mc *aMessageCenter) Push( msg *pubsub.Message, prefix string ) {

	mc.rmu.Lock()
	defer mc.rmu.Unlock()

	// TODO you need a firewall

	sender, err := mc.cvfs.Nodes().GetNodeByPeerId( msg.GetFrom().String() )

	if err != nil {
		return
	}

	if sender == nil {
		return
	}

	msgHash := crypto.Keccak256Hash(msg.Data)
	rmsg, exist := mc.reviewMsgMap.Load(msgHash)
	if !exist {

		rmsg = NewReviewMessage( msg.Data, sender, prefix, func(hash common.Hash) {
			mc.reviewMsgMap.Delete(hash)
			mc.totalCount --
		})

		mc.reviewMsgMap.Store(msgHash, rmsg)
		mc.totalCount ++

		return

	} else {

		m, ok := rmsg.(*ReviewMessage)
		if !ok {
			return
		}

		m.AddConfirmNode(sender)

		v, s, n := m.VoteInfo()
		if v > mc.threshold && s > mc.cnf.SuperNodeMin && n > mc.cnf.NodeTotalMin {

			switch m.MsgPrefix {

			case AMBlock.MessagePrefix:

				msg := &im.Block{}
				if err := proto.Unmarshal(m.Content, msg); err != nil {
					return
				}
				mc.replay <- msg

			case AMined.MessagePrefix:

				msg := &im.Minined{}
				if err := proto.Unmarshal(m.Content, msg); err != nil {
					return
				}
				mc.replay <- msg

			case ACInfo.MessagePrefix:

				msg := &im.ChainInfo{}
				if err := proto.Unmarshal(m.Content, msg); err != nil {
					return
				}
				mc.replay <- msg

			default:
				return

			}

			return
		}

	}

}

func (mc *aMessageCenter) TrustMessage() <- chan interface{} {
	return mc.replay
}

func (mc *aMessageCenter) PowerOn( ctx context.Context, chainID string, ind *core.IpfsNode ) {

	log.Info("AMC On")
	defer log.Info("AMC Off")

	var (

		err error

		mblockSuber *pubsub.Subscription
		batchSuber *pubsub.Subscription
		appendSuber *pubsub.Subscription

		awaiter sync.WaitGroup
	)

	msctx, mscancel := context.WithCancel(ctx)
	bsctx, bscancel := context.WithCancel(ctx)
	asctx, ascancel := context.WithCancel(ctx)

	// Create Subscribe
	mblockSuber, err = ind.PubSub.Subscribe( GetChannelTopics(chainID, MessageChannelMiningBlock) )
	if err != nil {
		goto ErrorReturn
	}
	defer mblockSuber.Cancel()

	batchSuber, err = ind.PubSub.Subscribe( GetChannelTopics(chainID, MessageChannelMined) )
	if err != nil {
		goto ErrorReturn
	}
	defer batchSuber.Cancel()

	appendSuber, err = ind.PubSub.Subscribe( GetChannelTopics(chainID, MessageChannelChainInfo) )
	if err != nil {
		goto ErrorReturn
	}
	defer appendSuber.Cancel()

	// MBlock Channel
	go func() {

		awaiter.Add(1)
		defer awaiter.Done()

		for {

			msg, err := mblockSuber.Next( msctx )
			if err != nil {
				return
			}

			mc.Push(msg, AMBlock.MessagePrefix)
		}

	}()

	// Batch Channel
	go func() {

		awaiter.Add(1)
		defer awaiter.Done()

		for {

			msg, err := batchSuber.Next( bsctx )
			if err != nil {
				return
			}

			mc.Push(msg, AMined.MessagePrefix)
		}

	}()

	// Appender Channel
	go func() {

		awaiter.Add(1)
		defer awaiter.Done()

		for {

			msg, err := appendSuber.Next( asctx )
			if err != nil {
				return
			}

			mc.Push(msg, ACInfo.MessagePrefix)
		}

	}()

	select {

	case <- ctx.Done():
		break

	case <- bsctx.Done():
		break

	case <- msctx.Done():
		break

	case <- asctx.Done():
		break
	}

	mscancel()
	bscancel()
	ascancel()

	awaiter.Wait()

	return

ErrorReturn:

	return
}

func (mc *aMessageCenter) PublishMessage( pmsg proto.Message, topic string ) error {

	cbs, err := proto.Marshal(pmsg)
	if err != nil {
		return err
	}

	if len(cbs) <= 0 {
		return ErrPublishedContentEmpty
	}

	return mc.ind.PubSub.Publish( topic, cbs )
}