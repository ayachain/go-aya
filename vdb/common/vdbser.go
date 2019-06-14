package common

import "github.com/syndtr/goleveldb/leveldb"


type VDBCacheServices interface {

	Close()

	MergerBatch() *leveldb.Batch
}

type VDBSerices interface {

	Shutdown() error

	NewCache() (VDBCacheServices, error)

	OpenTransaction() (*leveldb.Transaction, error)
}


