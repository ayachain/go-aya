package main

import (
	chancmd "github.com/ayachain/go-aya/chain/cmd"
	aappcmd "github.com/ayachain/go-aya/aapp/cmd"
	"github.com/ayachain/go-aya/keystore"
	walletcmd "github.com/ayachain/go-aya/keystore/cmd"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var ayacmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Display file status.",
	},
	Subcommands: map[string]*cmds.Command{
		"aapp" 		: 	aappcmd.AAppCMDS,
		"keystore" 	: 	walletcmd.WalletCMDS,
		"chain"		: chancmd.ChainCMDS,
	},

}

func DaemonAyaChaine() {
	keystore.Init("/Users/apple/.aya/keystore")
}
