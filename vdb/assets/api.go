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
	DRVer = byte(1)
	DBPATH = "/db/assets"
)

type AssetsAPI interface {
	AVdbComm.VDBSerices
	AssetsOf( key []byte ) ( *Assets, error )
	AvailBalanceMove( from, to []byte, v uint64 ) ( aftf, aftt *Assets, err error )
}