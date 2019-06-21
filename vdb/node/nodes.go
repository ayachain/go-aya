package node

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type aNodes struct {

	reader
	*mfs.Directory

	mfsstorage storage.Storage
	rawdb *leveldb.DB
}


func CreateServices( mdir *mfs.Directory, rdonly bool) Services {

	api := &aNodes{
		Directory:mdir,
	}

	api.rawdb, api.mfsstorage = AVdbComm.OpenExistedDB(mdir, DBPath, rdonly)

	return api
}

func (api *aNodes)  NewCache() (AVdbComm.VDBCacheServices, error) {
	return newCache( api.rawdb )
}

func (api *aNodes)  OpenTransaction() (*leveldb.Transaction, error) {

	tx, err := api.rawdb.OpenTransaction()

	if err != nil {
		return nil, err
	}

	return tx, nil
}


func (api *aNodes)  Shutdown() error {

	_ = api.rawdb.Close()
	_ = api.mfsstorage.Close()

	return api.Flush()
}


func (api *aNodes) GetNodeByPeerId( peerId string ) (*Node, error) {

	bs, err := api.rawdb.Get( []byte(peerId), nil )
	if err != nil {
		return nil, err
	}

	nd := &Node{}

	if err := nd.Decode(bs); err != nil {
		return nil, err
	}

	return nd, nil
}