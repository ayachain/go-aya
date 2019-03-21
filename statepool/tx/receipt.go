package tx

import "encoding/json"

type TxReceipt struct {
	BlockIndex uint64
	TxHash	string
	Response interface{}
}


func (txr *TxReceipt) MarshalJson() (bs []byte, err error) {

	bs, err = json.Marshal(txr)

	return bs,err
}