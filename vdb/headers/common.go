package headers

import "github.com/ipfs/go-cid"

func HeadersToCids( hs []*Header ) []cid.Cid {

	var cids []cid.Cid

	for _, v := range hs {
		cids = append(cids, v.Cid)
	}

	return cids
}