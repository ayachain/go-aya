package receipt

import (
	"encoding/json"
	ArawdbComm "github.com/ayachain/go-aya/vdb/common"
)

type Receipt struct {

	ArawdbComm.RawDBCoder 	`json:"-"`

	Stat int				`json:"Stat, omitempty"`
	Message string			`json:"Msg, omitempty"`
	Event 	[]string		`json:"Events, omitempty"`
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


func ConfirmReceipt( msg string, events []string ) *Receipt {

	return &Receipt{
		Stat:0,
		Message:msg,
		Event:events,
	}

}

func ExpectedReceipt( msg string, events []string ) *Receipt {

	return &Receipt{
		Stat:-1,
		Message:msg,
		Event:events,
	}
}