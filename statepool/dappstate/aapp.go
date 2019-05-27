package dappstate

import (
	"context"
	"encoding/json"
	"github.com/ipfs/go-ipfs-api"
	"io/ioutil"
	"strings"
)

type Aapp struct {
	MasterNodes		[]string
	StateNames		[]string
}

type ipnskey struct {
	Name	string
	Id		string
}

type ipnskeylist struct {
	Keys	[]ipnskey
}

func (a *Aapp) GetUnionStateNames() []string {

	reqb, err := shell.NewLocalShell().Request("key/list").Option("l",true).Send(context.Background())

	_, cancel := context.WithCancel(context.Background())

	defer cancel()

	//iface.Key().Path()

	if err != nil {
		return nil
	}

	reqc, err := ioutil.ReadAll(reqb.Output)

	if err != nil {
		return nil
	}

	list := &ipnskeylist{}

	if err := json.Unmarshal(reqc, list); err != nil {
		return nil
	}

	var ret []string

	for _, v1 := range list.Keys {

		for _, v2 := range a.StateNames {

			if strings.EqualFold(v1.Id, v2) {
				ret = append(ret, v1.Id)
				break
			}

		}

	}

	return ret

}