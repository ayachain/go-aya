package aapp

import (
	"errors"
	iface "github.com/ipfs/interface-go-ipfs-core"
)

var Manager = &mrg{}

type mrg struct {
	aapps map[string]*aapp
}

func ( m *mrg ) List() []string {

	var l []string

	for k, _ := range m.aapps {
		l = append(l, k)
	}

	return l
}

func ( m *mrg ) Load( aappns string, api iface.CoreAPI ) ( ap *aapp, err error ) {

	_, isexist := m.aapps[aappns]

	if isexist {
		return nil, errors.New("AApp is already exist in node.")
	}

	ap, err = NewAApp(aappns, api)

	if err != nil {
		return nil, err
	} else {
		m.aapps[aappns] = ap
	}

	return ap, nil

}

