package assets

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
)

const (
	//Default assets record version code
	DRVer 				= byte(1)
	DBPath 				= "/assets"
)

type reader interface {
	AssetsOf( addr EComm.Address, idx... *indexes.Index ) ( *Assets, error )
}

type writer interface {
	Put( addr EComm.Address, ast *Assets )
}

type Services interface {
	AVdbComm.VDBSerices
	reader
}


type MergeWriter interface {
	AVdbComm.VDBCacheServices
	reader
	writer
}