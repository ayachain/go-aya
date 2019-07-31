package cmd

import cmds "github.com/ipfs/go-ipfs-cmds"

var NodeCMDS = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "AyaChain network node commands.",
	},
	Subcommands: map[string]*cmds.Command{
		"list"	 		: listCMD,
		"online"		: cstatCMD,
	},

}