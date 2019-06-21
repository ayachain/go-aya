package node

import (
	"bytes"
	"encoding/json"
	"errors"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-cid"
)

type NodeType string

const (
	NodeTypeSuper 	NodeType = "Super"
	NodeTypeMaster 	NodeType = "Master"
)

type Node struct {
	
	AVdbComm.RawDBCoder     `json:"-"`
	
	Type NodeType 			`json:"Type"`

	PeerID string			`json:"PeerID"`

	Owner EComm.Address		`json:"Owner"`
	
	Sig []byte				`json:"Sig"`

}