package msgcenter

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
)

type MessageChannel string

const (
	Version 									= "AyaMessageCenter V0.0.1"
	MessageChannelMiningBlock	MessageChannel	= "MiningBlock"
	MessageChannelMined			MessageChannel	= "Mined"
	MessageChannelChainInfo		MessageChannel	= "ChainInfo"
)

func GetChannelTopics( chainID string, c MessageChannel ) string {
	return Version + " " + crypto.Keccak256Hash([]byte(fmt.Sprintf("%v%v", chainID, c))).String()
}