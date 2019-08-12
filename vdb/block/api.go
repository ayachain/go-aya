package block

import AVdbComm "github.com/ayachain/go-aya/vdb/common"

const DBPath = "/blocks"

type reader interface {

	GetBlocks( hashOrIndex...interface{} ) ([]*Block, error)

	GetLatestPosBlockIndex() uint64

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


type Caches interface {
	AVdbComm.VDBCacheServices
	reader
	writer
}