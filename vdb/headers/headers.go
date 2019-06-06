package headers

import (
	"encoding/binary"
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-cid"
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

	c, err := cid.Cast(bs)
	if err != nil {
		return nil, err
	}

	return &Header{Cid:c}, nil
}

func (hds *aHeaders) AppendHeaders( header... *Header ) error {

	lindex := hds.LatestHeaderIndex()

	wbc := &leveldb.Batch{}
	for _, v := range header {
		lindex ++
		wbc.Put( common.BigEndianBytes(lindex), v.Encode() )
	}

	wbc.Put([]byte(latestHeaderNumKey), common.BigEndianBytes(lindex) )

	return hds.rawdb.Write(wbc, nil)
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