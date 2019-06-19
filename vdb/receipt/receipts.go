package receipt

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

type aReceipt struct {

	Services
	*mfs.Directory

	mfsstorage storage.Storage
	rawdb *leveldb.DB
	RWLocker sync.RWMutex
}

func CreateServices( mdir *mfs.Directory, rdonly bool ) Services {

	api := &aReceipt{
		Directory:mdir,
	}

	api.rawdb, api.mfsstorage = AVdbComm.OpenExistedDB(mdir, DBPath, rdonly)

	return api
}

func (r *aReceipt) HasTransactionReceipt( txhs EComm.Hash ) bool {

	st := append(txhs.Bytes(), AVdbComm.BigEndianBytes(0)... )
	ed := append(txhs.Bytes(), AVdbComm.BigEndianBytes((uint64(1) << 63) -1)... )

	s, err := r.rawdb.SizeOf([]util.Range{{Start:st,Limit:ed}})
	if err != nil {
		return false
	}

	return s.Sum() > 0
}

func (r *aReceipt) GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error) {

	if !r.HasTransactionReceipt(txhs) {
		return nil, leveldb.ErrNotFound
	}

	st := append(txhs.Bytes(), AVdbComm.BigEndianBytes(0)... )
	ed := append(txhs.Bytes(), AVdbComm.BigEndianBytes((uint64(1) << 63) -1)... )

	it := r.rawdb.NewIterator(&util.Range{Start:st,Limit:ed}, nil)

	if !it.Next() {
		return nil, leveldb.ErrNotFound
	}

	rp := &Receipt{}

	err := rp.Decode( it.Value() )
	if err != nil {
		return nil, err
	}

	return rp, nil
}

func (r *aReceipt) Close() {

	_ = r.rawdb.Close()

	_ = r.mfsstorage.Close()

	_ = r.Flush()

}

func (r *aReceipt) NewCache() (AVdbComm.VDBCacheServices, error) {

	return newCache( r.rawdb )

}

func (r *aReceipt) OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := r.rawdb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (r *aReceipt) Shutdown() error {

	_ = r.rawdb.Close()
	_ = r.mfsstorage.Close()

	return r.Flush()
}