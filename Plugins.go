package main

import (
	"fmt"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
)

type AyaPlugin struct{}

var AyaPluginType = "ayaplugin"

func (*AyaPlugin) Name() string {
	return "ds-delaystore"
}

// Version returns the plugin's version, satisfying the plugin.Plugin interface.
func (*AyaPlugin) Version() string {
	return "0.1.0"
}

// Init initializes plugin, satisfying the plugin.Plugin interface. Put any
// initialization logic here.
func (*AyaPlugin) Init() error {
	return nil
}

func (*AyaPlugin) Start(api coreiface.CoreAPI) error {
	fmt.Println("Hello!")
	return nil
}

func (*AyaPlugin) Close() error {
	fmt.Println("Goodbye!")
	return nil
}