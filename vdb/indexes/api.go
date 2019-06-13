package indexes

import (
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
)

type IndexesAPI interface {

	PutIndex( index *Index ) error

	PutIndexBy( num uint64, bhash EComm.Hash, cid cid.Cid ) error

	GetIndex( blockNumber uint64 ) (*Index, error)

	Close() error
}