package node

import (
	"encoding/json"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"time"
)

type NodeType string

const (
	NodeTypeSuper 	NodeType = "Super"
	NodeTypeMaster 	NodeType = "Master"
)

type PingRet struct {
	Node *Node
	UTime int64
	RTT time.Duration
}

type Node struct {
	
	AVdbComm.RawDBCoder     `json:"-"`
	
	Type NodeType 			`json:"Type,omitempty"`
	
	Votes uint64			`json:"Votes,omitempty"`

	PeerID string			`json:"PeerID,omitempty"`

	Owner EComm.Address		`json:"Owner,omitempty"`
	
	Sig []byte				`json:"Sig,omitempty"`

}

func (nd *Node) Encode() []byte {

	bs, err := json.Marshal(nd)

	if err != nil {
		return nil
	}

	return bs
}


func (nd *Node) Decode(bs []byte) error {
	return json.Unmarshal(bs, nd)
}

