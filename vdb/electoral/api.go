package electoral

import AVdbComm "github.com/ayachain/go-aya/vdb/common"

const DBPath = "/electoral"

type reader interface {
	PackerOf(index uint64) (*Electoral, error)
}

type writer interface {
	AppendPacker( packer *Electoral ) error
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