package act

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type TxRspAct struct {
	BaseAct
	BlockHash	string
	ResultState	string
}

func NewTxRspAct(dappns string, bhash string, rhash string) BaseActInf {
	return &TxRspAct{BaseAct{TStr:"TxRspAct", DPath:dappns}, bhash, rhash}
}

func (act *TxRspAct) EncodeToHex() (hex string, err error) {

	bs, err := json.Marshal(act)

	if err != nil {
		return "",err
	}

	return hexutil.Encode(bs),nil
}

func (act *TxRspAct) DecodeFromHex(hex string) (err error) {

	bs, err := hexutil.Decode(hex)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bs, act); err != nil {
		return err
	}

	return nil

}