package block

import (
	"github.com/ayachain/go-aya/vdb/headers"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
)

type aBlocks struct {
	BlocksAPI
	ind *core.IpfsNode
	headAPI headers.HeadersAPI
}

func CreateBlocksAPI(ind *core.IpfsNode, hapi headers.HeadersAPI) BlocksAPI {

	api := &aBlocks{
		ind:ind,
		headAPI:hapi,
	}

	return api

}

func (blks *aBlocks) GetBlocks ( iorc... interface{} ) ([]*Block, error) {

	var cids []cid.Cid

	for _, v := range iorc {

		switch v.(type) {

		case uint64:

			hd, err := blks.headAPI.HeaderOf(v.(uint64))
			if err != nil {
				return nil, err
			}
			cids = append(cids, hd.Cid)

		case cid.Cid:
			cids = append(cids, v.(cid.Cid))
		}

	}

	return GetBlocks(blks.ind, cids... )
}