package msgcenter

import (
	ANode "github.com/ayachain/go-aya/vdb/node"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"time"
)

type ReviewMessage struct {

	Content []byte

	Hash EComm.Hash

	Nodes []*ANode.Node
}


func NewReviewMessage( contnet []byte, sender *ANode.Node, dfunc func(hash EComm.Hash) ) *ReviewMessage {

	message := &ReviewMessage{
		Content:contnet,
		Hash:crypto.Keccak256Hash(contnet),
		Nodes:[]*ANode.Node{sender},
	}

	go func() {

		etimer := time.NewTimer(time.Second * 32)

		defer etimer.Stop()

		<- etimer.C

		dfunc( message.Hash )

	}()

	return message
}


func (msg *ReviewMessage) AddConfirmNode( confirmer *ANode.Node ) {

	for _, n := range msg.Nodes {

		if n.PeerID == confirmer.PeerID {
			return
		}

	}

	msg.Nodes = append(msg.Nodes, confirmer)
}

func (msg *ReviewMessage) VoteInfo() (votes uint64, scount uint, ncount uint) {

	for _, n := range msg.Nodes {

		if n.Type == ANode.NodeTypeSuper {
			scount ++
		}
		votes += n.Votes
	}

	return votes, scount, uint(len(msg.Nodes))
}