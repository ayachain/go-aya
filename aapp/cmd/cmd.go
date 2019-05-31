package cmd

import (
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var AAppCMDS = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "AyaChain AApp Commands.",
	},
	Subcommands: map[string]*cmds.Command{
		"daemon" 	: 	daemonCmd,
		"list"	 	: 	listCmd,
		"shutdown"	:	shutdownCmd,
		"perfrom"	:	perfromCmd,
	},

}