package act

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type BaseAct struct {
	TStr	string
	DPath	string
}

func NewBaseAct(dpath string) (act* BaseAct) {
	return &BaseAct{"BaseAct", dpath}
}

func (act *BaseAct) EncodeToHex() (hex string, err error) {

	bs, err := json.Marshal(act)

	if err != nil {
		return "",err
	}

	return hexutil.Encode(bs),nil
}

func (act *BaseAct) DecodeFromHex(hex string) (err error) {

	bs, err := hexutil.Decode(hex)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bs, act); err != nil {
		return err
	}

	return nil

}