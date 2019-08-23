package cmd

import (
	"github.com/ipfs/go-ipfs-cmds"
)

var ChainCMDS = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "AyaChain Chain Commands.",
	},
	Subcommands: map[string]*cmds.Command{
		"config"		:	configCmd,
		"connect" 		: 	connCmd,
		"disconnect" 	: 	dissConnectCmd,
	},

}
