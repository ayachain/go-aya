package cmd

import (
	"github.com/ipfs/go-ipfs-cmds"
)

var BlockCMDS = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "AyaChain Block Commands.",
	},
	Subcommands: map[string]*cmds.Command{
		"get" 				: 	getCmd,
	},

}
