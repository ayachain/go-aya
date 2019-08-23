package assets

import (
	"github.com/ayachain/go-aya/vdb/im"
	EComm "github.com/ethereum/go-ethereum/common"
)

type SortAssets struct {
	*im.Assets
	Address		EComm.Address
}

func NewAssets( avail, vote, locked uint64 ) *im.Assets {

	return &im.Assets{
		Avail:avail,
		Vote:vote,
		Locked:locked,
	}

}