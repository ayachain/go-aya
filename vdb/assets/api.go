package assets

import (
	"errors"
)

var (
	notEnoughError = errors.New("not enough balance")
)

const (
	//Default assets record version code
	DRVer = byte(1)
	availDBPath = "/db/rawdb"
)

type AssetsAPI interface {
	AssetsOf( key []byte ) ( *Assets, error )
	AvailBalanceMove( from, to []byte, v uint64 ) ( aftf, aftt *Assets, err error )
}