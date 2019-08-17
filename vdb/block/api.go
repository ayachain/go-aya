package block

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
)

const DBPath = "/blocks"

type reader interface {

	GetLatestBlock( ) (*Block, error)

	GetBlocks( hashOrIndex...interface{} ) ([]*Block, error)

	GetLatestPosBlockIndex(idx... *indexes.Index) uint64
}

type writer interface {

	AppendBlocks( blocks...*Block )

	WriteGenBlock( gen *GenBlock )

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