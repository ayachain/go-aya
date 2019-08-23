package assets

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
)

const DBPath = "/assets"

type reader interface {
	AssetsOf( addr EComm.Address, idx... *indexes.Index ) ( *im.Assets, error )
}

type writer interface {
	Put( addr EComm.Address, ast *im.Assets )
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