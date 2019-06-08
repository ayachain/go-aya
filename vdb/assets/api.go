package assets

import (
	"errors"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
)

var (
	notEnoughError = errors.New("not enough balance")
)

const (
	//Default assets record version code
	DRVer 				= byte(1)
	DBPATH 				= "/db/assets"
	DBTopIndexPrefix 	= "Top_"
)

type AssetsAPI interface {

	AVdbComm.VDBSerices

	VotingCountOf( key []byte ) ( uint64, error )

	AssetsOf( key []byte ) ( *Assets, error )

	AvailBalanceMove( from, to []byte, v uint64 ) ( aftf, aftt *Assets, err error )

	GetLockedTop100() ( []*SortAssets, error )
}