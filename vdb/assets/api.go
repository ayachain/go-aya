package assets

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
)

const (
	//Default assets record version code
	DRVer 				= byte(1)
	DBPATH 				= "/assets"
)

type reader interface {
	AssetsOf( addr EComm.Address ) ( *Assets, error )
}

type writer interface {

	PutNewAssets( addr EComm.Address, ast *Assets )

	Put( addr EComm.Address, ast *Assets )
}


type Services interface {
	AVdbComm.VDBSerices
	reader
}


type Caches interface {
	AVdbComm.VDBCacheServices
	reader
	writer
}