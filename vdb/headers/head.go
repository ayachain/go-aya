package headers

import (
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-cid"
)

type Header struct {
	common.RawDBCoder
	cid.Cid
}


func (h *Header) Encode() []byte {
	return h.Cid.Bytes()
}

func (h *Header) Decode(bs []byte) error {

	c, err := cid.Cast(bs)
	if err != nil {
		return err
	}

	h.Cid = c

	return nil
}