package msgcenter

import (
	"bytes"
	"github.com/ayachain/go-aya/vdb/im"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"time"
)

type ReviewMessage struct {

	Content []byte

	Hash EComm.Hash

	Nodes []*im.Node

	MsgPrefix string
}

func NewReviewMessage( content []byte, sender *im.Node, prefix string, dfunc func(hash EComm.Hash) ) *ReviewMessage {

	message := &ReviewMessage{
		Content:content,
		Hash:crypto.Keccak256Hash(content),
		Nodes:[]*im.Node{sender},
		MsgPrefix:prefix,
	}

	go func() {

		etimer := time.NewTimer(time.Second * 32)

		defer etimer.Stop()

		<- etimer.C

		dfunc( message.Hash )

	}()

	return message
}

func (msg *ReviewMessage) AddConfirmNode( confirmer *im.Node ) {

	for _, n := range msg.Nodes {

		if n.PeerID == confirmer.PeerID {
			return
		}

	}

	msg.Nodes = append(msg.Nodes, confirmer)
}

func (msg *ReviewMessage) VoteInfo() (votes uint64, scount uint, ncount uint) {

	for _, n := range msg.Nodes {

		if n.Type == im.NodeType_Super {
			scount ++
		}
		votes += n.Votes
	}

	return votes, scount, uint(len(msg.Nodes))
}

func (msg *ReviewMessage) Description() string {
	return  string(bytes.ToUpper(msg.Content[:1]))
}