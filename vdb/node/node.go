package node

import (
	"encoding/json"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
)

type NodeType string

const (
	NodeTypeSuper 	NodeType = "Super"
	NodeTypeMaster 	NodeType = "Master"
)

type Node struct {
	
	AVdbComm.RawDBCoder     `json:"-"`
	
	Type NodeType 			`json:"Type"`
	
	Votes uint64			`json:"Votes"`

	PeerID string			`json:"PeerID"`

	Owner EComm.Address		`json:"Owner"`
	
	Sig []byte				`json:"Sig"`

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

