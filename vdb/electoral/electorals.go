package electoral

import (
	ADB "github.com/ayachain/go-aya-alvm-adb"
	"github.com/ayachain/go-aya/logs"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/whyrusleeping/go-logging"
	"sync"
)

var log = logging.MustGetLogger(logs.AVDBServices_Electorals)

type aElectorals struct {

	Services
	reader
	*mfs.Directory

	ldb *leveldb.DB
	mfsstorage *ADB.MFSStorage
	dbSnapshot *leveldb.Snapshot
	snLock sync.RWMutex
}

func CreateServices( mdir *mfs.Directory ) Services {

	var err error

	api := &aElectorals{
		Directory:mdir,
	}

	api.ldb, api.mfsstorage = AVdbComm.OpenExistedDB( mdir, DBPath )

	api.dbSnapshot, err = api.ldb.GetSnapshot()

	if err != nil {
		_ = api.ldb.Close()
		log.Error(err)
		return nil
	}

	return api
}

func (api *aElectorals) PackerOf(index uint64) (*Electoral, error) {

	bs, err := api.ldb.Get( AVdbComm.BigEndianBytes(index), nil )

	if err != nil {
		return nil, err
	}

	ele := &Electoral{}

	if err = ele.Decode(bs); err != nil {
		return nil, err
	}

	return ele, nil
}


func (api *aElectorals) NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( api.dbSnapshot )
}

func (api *aElectorals) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := api.ldb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (api *aElectorals) Shutdown() error {

	api.snLock.Lock()
	defer api.snLock.Unlock()

	if api.dbSnapshot != nil {
		api.dbSnapshot.Release()
	}

	if err := api.ldb.Close(); err != nil {
		return err
	}

	return nil
}

func (api *aElectorals) Close() {
	_ = api.Shutdown()
}

func (api *aElectorals) UpdateSnapshot() error {

	api.snLock.Lock()
	defer api.snLock.Unlock()

	if api.dbSnapshot != nil {
		api.dbSnapshot.Release()
	}

	var err error
	api.dbSnapshot, err = api.ldb.GetSnapshot()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (api *aElectorals) SyncCache() error {

	if err := api.ldb.CompactRange(util.Range{nil,nil}); err != nil {
		log.Error(err)
	}

	return api.mfsstorage.Flush()
}
