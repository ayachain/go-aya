package act

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type PerfromAct struct {
	BaseAct
	Method		string
	Parmas		string
}

func NewPerfromAct(dpath string, method string, parmas string) BaseActInf {

	r := &PerfromAct {
		BaseAct:BaseAct {
			TStr:"PerfromAct",
			DPath:dpath,
		},
		Method:method,
		Parmas:parmas,
	}

	return r
}


func (act *PerfromAct) EncodeToHex() (hex string, err error) {

	bs, err := json.Marshal(act)

	if err != nil {
		return "",err
	}

	return hexutil.Encode(bs),nil
}

func (act *PerfromAct) DecodeFromHex(hex string) (err error) {

	bs, err := hexutil.Decode(hex)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bs, act); err != nil {
		return err
	}

	return nil

}