package indexes

import (
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
)

type IndexesServices interface {

	GetIndex( blockNumber uint64 ) (*Index, error)

	GetLatest() (*Index, error)

	PutIndex( index *Index ) (cid.Cid, error)

	PutIndexBy( num uint64, bhash EComm.Hash, cid cid.Cid ) (cid.Cid, error)

	Close() error

	Flush() error
}
