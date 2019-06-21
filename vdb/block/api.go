package block

import AVdbComm "github.com/ayachain/go-aya/vdb/common"

const DBPath = "/db/blocks"

type reader interface {

	GetBlocks( hashOrIndex...interface{} ) ([]*Block, error)

}

type writer interface {

	AppendBlocks( blocks...*Block )

	WriteGenBlock( gen *GenBlock )
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