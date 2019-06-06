package common

import (
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

type VDBSerices interface {

	Close()

	DBKey() string

	OpenVDBTransaction() (*leveldb.Transaction, *sync.RWMutex, error)

}