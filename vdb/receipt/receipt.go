package receipt

import (
	"encoding/json"
	ArawdbComm "github.com/ayachain/go-aya/vdb/common"
)

type Receipt struct {
	ArawdbComm.RawDBCoder
	Code int16		`json:"code"`
	Body []byte		`json:"body"`
	Event []string	`json:"events"`
}

func (r *Receipt) Encode() []byte {

	bs, err := json.Marshal(r)

	if err != nil {
		return nil
	} else {
		return bs
	}

}

func (r *Receipt) Decode(bs []byte) error {
	return json.Unmarshal(bs, r)
}

