package tx

import (
	"encoding/json"
)

type TxReceipt struct {
	BlockIndex 	uint64
	Tx			Tx
	TxHash		string
	Status		string
	Response 	interface{}
}


func (txr *TxReceipt) MarshalJson() (bs []byte, err error) {

	bs, err = json.Marshal(txr)

	return bs,err
}