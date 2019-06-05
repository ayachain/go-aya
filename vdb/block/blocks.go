package block

import (
	"errors"
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/headers"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
)

type aBlocks struct {

	BlocksAPI
	*mfs.Directory

	headAPI headers.HeadersAPI
	rawdb *leveldb.DB
}

func CreateServices( mdir *mfs.Directory, hapi headers.HeadersAPI) BlocksAPI {

	api := &aBlocks{
		Directory:mdir,
		headAPI:hapi,
	}

	api.rawdb = common.OpenExistedDB(mdir, DBPath)

	return api
}

func (blks *aBlocks) GetBlocks ( iorc... interface{} ) ([]*Block, error) {

	var blist []*Block

	for _, v := range iorc {

		var bcid cid.Cid

		switch v.(type) {

		case uint64:
			hd, err := blks.headAPI.HeaderOf(v.(uint64))
			if err != nil {
				return nil, err
			}
			bcid = hd.Cid

		case cid.Cid:
			bcid = v.(cid.Cid)

		default:
			return nil, errors.New("input params must be a index(uint64) or cid object")
		}

		dbval, err := blks.rawdb.Get( bcid.Bytes(), nil )
		if err != nil {
			return nil, err
		}

		subBlock := &Block{}
		if err := subBlock.Decode(dbval); err != nil {
			return nil, err
		}

		blist = append(blist, subBlock)
	}

	return blist, nil
}