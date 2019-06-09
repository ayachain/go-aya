package headers

import (
	"encoding/json"
	"github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
)

type Header struct {
	common.RawDBCoder
	BlockIndex uint64 	`json:"I"`
	Hash EComm.Hash		`json:"H"`
}

func (h *Header) Encode() []byte {

	bs, err := json.Marshal(h)

	if err != nil {
		return nil
	}

	return bs
}

func (h *Header) Decode(bs []byte) error {
	return json.Unmarshal(bs, h)
}