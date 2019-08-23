package block

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/indexes"
)

const DBPath = "/blocks"

type reader interface {

	GetLatestBlock( ) (*im.Block, error)

	GetBlocks( hashOrIndex...interface{} ) ([]*im.Block, error)

	GetLatestPosBlockIndex(idx... *indexes.Index) uint64
}

type writer interface {

	AppendBlocks( blocks...*im.Block )

	WriteGenBlock( gen *im.GenBlock )

	SetLatestPosBlockIndex( idx uint64 )
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