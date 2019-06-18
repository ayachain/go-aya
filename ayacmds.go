package main

import (
	aappcmd "github.com/ayachain/go-aya/aapp/cmd"
	"github.com/ayachain/go-aya/chain"
	chancmd "github.com/ayachain/go-aya/chain/cmd"
	"github.com/ayachain/go-aya/keystore"
	keystorecmd "github.com/ayachain/go-aya/keystore/cmd"
	txcmd "github.com/ayachain/go-aya/tx/cmd"
	walletcmd "github.com/ayachain/go-aya/wallet/cmd"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core"
)

var ayacmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Display file status.",
	},
	Subcommands: map[string]*cmds.Command{
		"aapp" 		: 	aappcmd.AAppCMDS,
		"keystore" 	: 	keystorecmd.KeystoreCMDS,
		"chain"		: 	chancmd.ChainCMDS,
		"wallet"	:	walletcmd.WalletCMDS,
		"tx"		:	txcmd.TxCMDS,
	},

}


func DaemonAyaChain( ind *core.IpfsNode ) {
	keystore.Init("/Users/apple/.aya/keystore", ind)
}


func ShutdownAyaChain() <- chan bool {
	return chain.DisconnectionAll()
}