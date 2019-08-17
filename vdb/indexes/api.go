package indexes

import (
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
)

type IndexesServices interface {

	GetLatest() ( *Index, error )

	SyncToCID( fullCID cid.Cid ) error

	PutIndex( index *Index ) (cid.Cid, error)

	GetIndex( blockNumber uint64 ) (*Index, error)

	PutIndexBy( num uint64, bhash EComm.Hash, cid cid.Cid ) (cid.Cid, error)

	Close() error
}