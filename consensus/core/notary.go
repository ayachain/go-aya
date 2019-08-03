//
// Notary interface is one of the most important failures in consensus mechanism.
// You can compare it to the same person on different nodes. Their working mode
// is identical. So long as the output is the same, the notary knows the steps of
// work, so that the same final result can be obtained through the same logic in
// different nodes, and different consensus can be supported by realizing notary
// interface.
//
// Mechanisms, such as POS, POW, DPOS and so on, will provide the basic access
// interface of the source data, fully publicizing all the rights to write and
// read to the notary.
//
// When reading and viewing data, notaries are not required to participate. Notaries
// only participate in supervision and unified logic when new data is written to
// block chain database, such as when receiving a new block of data.
//
package core

import (
	AGroup "github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ayachain/go-aya/vdb"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type NotaryMessageType string

const (

	NotaryMessageTransaction 	NotaryMessageType = "tx"
	NotaryMessageMiningBlock 	NotaryMessageType = "mblock"
	NotaryMessageConfirmBlock 	NotaryMessageType = "cblock"
	NotaryMessageMinedRet		NotaryMessageType = "minedret"
	NotaryMessageChainInfo		NotaryMessageType = "info"

)

type Notary interface {

	MiningBlock( block *AMsgMBlock.MBlock, cvfs vdb.CacheCVFS, txs []*ATx.Transaction ) (*AGroup.TaskBatchGroup, error)

	TrustOrNot( msg *pubsub.Message, mtype NotaryMessageType, cvfs vdb.CVFS ) <- chan bool

 	NewBlockHasConfirm()
}