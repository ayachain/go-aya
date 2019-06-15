package main

import (
	aappcmd "github.com/ayachain/go-aya/aapp/cmd"
	chancmd "github.com/ayachain/go-aya/chain/cmd"
	"github.com/ayachain/go-aya/keystore"
	keystorecmd "github.com/ayachain/go-aya/keystore/cmd"
	txcmd "github.com/ayachain/go-aya/tx/cmd"
	walletcmd "github.com/ayachain/go-aya/wallet/cmd"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core"
	"github.com/whyrusleeping/go-logging"
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


var format = logging.MustStringFormatter(
	`%{color}%{time:05:11:22} %{shortfunc} : %{level} %{color:reset} : %{message}`,
)

func DaemonAyaChaine( ind *core.IpfsNode ) {

	//backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	//backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	//backend2Formatter := logging.NewBackendFormatter(backend2, format)
	//backend1Leveled := logging.AddModuleLevel(backend1)
	//backend1Leveled.SetLevel(logging.ERROR, "")
	//logging.SetBackend(backend1Leveled, backend2Formatter)

	keystore.Init("/Users/apple/.aya/keystore", ind)
}
