package swarm

import (
	"errors"
	"fmt"
	"github.com/ipfs/interface-go-ipfs-core"
)

type swarm struct {
	swarms map[string]*channel
}

func ( m *swarm ) ChannelList( ) []string {

	var list []string

	for k, _ := range m.swarms {
		list = append(list, k)
	}

	return list
	
}

func ( m *swarm) RegisterChannel( aappns string, handle func(_ iface.PubSubMessage) ) ( c *channel, err error ) {

	_, isexist := m.swarms[aappns]

	if isexist {
		return nil, errors.New(fmt.Sprintf(`RegisterChannel：标识为"%v"的信道已存在，请勿重复注册信道.`, aappns) )
	}

	c = NewChannel(aappns, handle)

	if err := c.Open(); err != nil {
		return nil, err
	}

	m.swarms[aappns] = c

	return m.swarms[aappns] ,nil
}

func ( m *swarm) UnRegisterChannel( aappns string ) error {

	c, isexist := m.swarms[aappns]

	if !isexist {
		return errors.New(fmt.Sprintf(`UnRegisterChannel：标识为"%v"的信道不存在`, aappns))
	}

	if err := c.Close(); err != nil {
		return err
	}

	delete(m.swarms, aappns)

	return nil
}