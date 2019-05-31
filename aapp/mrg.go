package aapp

import (
	"errors"
	"fmt"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/interface-go-ipfs-core"
)

var Manager = &mrg{aapps: map[string]*aapp{}}

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

func ( m *mrg ) Load( aappns string, api iface.CoreAPI, ind *core.IpfsNode ) ( ap *aapp, err error ) {

	_, isexist := m.aapps[aappns]
	if isexist {
		return nil, errors.New("AApp is already exist in node.")
	}

	ap, err = NewAApp(aappns, api, ind)
	if err != nil {
		return nil, err
	} else {
		m.aapps[aappns] = ap
	}

	return ap, nil
}

func ( m *mrg ) Shutdown( aappns string ) error {

	ap, ok := m.aapps[aappns]
	if !ok {
		return fmt.Errorf("not find AApp : %v", aappns)
	}

	ap.Shutdown()

	delete(m.aapps, aappns)

	return nil
}

func ( m *mrg ) AAppOf( aappns string ) *aapp {

	v, isexist := m.aapps[aappns]

	if isexist {
		return v
	} else {
		return nil
	}

}
