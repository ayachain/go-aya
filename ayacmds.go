package main

import (
	aappcmd "github.com/ayachain/go-aya/aapp/cmd"
	walletcmd "github.com/ayachain/go-aya/keystore/cmd"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var ayacmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Display file status.",
	},
	Subcommands: map[string]*cmds.Command{
		"aapp" 		: 	aappcmd.AAppCMDS,
		"wallet" 	: 	walletcmd.WalletCMDS,
	},

}
