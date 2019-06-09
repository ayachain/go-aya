package headers

import (
	"encoding/binary"
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)


type aHeaders struct {
	HeadersAPI
	*mfs.Directory

	RWLocker sync.RWMutex

	rawdb *leveldb.DB
}

func (hds *aHeaders) LatestHeaderIndex() uint64 {

	bs, err := hds.rawdb.Get([]byte(latestHeaderNumKey), nil)

	if err != nil {
		return 0
	}

	return binary.BigEndian.Uint64(bs)
}

func (hds *aHeaders) HeaderOf( index uint64 ) (*Header, error) {

	bs, err := hds.rawdb.Get(common.BigEndianBytes(index), nil)
	if err != nil {
		return nil, err
	}

	hd := &Header{}
	if err := hd.Decode(bs); err != nil {
		return nil, err
	}

	return hd, nil
}

func (hds *aHeaders) AppendHeaders( group *AWork.TaskBatchGroup, header... *Header) error {

	lindex := hds.LatestHeaderIndex()

	for _, v := range header {
		lindex ++
		group.Put(DBPATH, common.BigEndianBytes(lindex), v.Encode())
	}

	group.Put(DBPATH, []byte(latestHeaderNumKey), common.BigEndianBytes(lindex) )

	return nil
}

func (txs *aHeaders) DBKey()	string {
	return DBPATH
}

func CreateServices( mdir *mfs.Directory ) HeadersAPI {

	api := &aHeaders{
		Directory:mdir,
	}

	api.rawdb = common.OpenExistedDB(mdir, DBPATH)

	return api
}

func (api *aHeaders) OpenVDBTransaction() (*leveldb.Transaction, *sync.RWMutex, error) {

	tx, err := api.rawdb.OpenTransaction()
	if err != nil {
		return nil, nil, err
	}

	return tx, &api.RWLocker, nil
}


func (api *aHeaders) Close() {

	api.RWLocker.Lock()
	defer api.RWLocker.Unlock()

	_ = api.rawdb.Close()

}