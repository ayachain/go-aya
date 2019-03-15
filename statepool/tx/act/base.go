package act

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type BaseActInf interface {
	EncodeToHex() (hex string, err error)
	DecodeFromHex(hex string) (err error)
}

type BaseAct struct {
	BaseActInf
	TStr	string
	DPath	string
}

func NewBaseAct(dpath string) BaseActInf {
	return &BaseAct{TStr:"BaseAct", DPath:dpath}
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