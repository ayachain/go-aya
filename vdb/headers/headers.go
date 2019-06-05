package headers

import (
	"encoding/binary"
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
)


type aHeaders struct {
	HeadersAPI
	*mfs.Root
	rawdb *leveldb.DB
	ind *core.IpfsNode
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


func CreateHeadersAPI( rootref *mfs.Root ) HeadersAPI {

	api := &aHeaders{
		Root:rootref,
	}

	api.rawdb = common.OpenStandardDB(rootref, headersDBPath)

	return api
}