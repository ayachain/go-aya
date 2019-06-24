package main

import (
	AappCMD "github.com/ayachain/go-aya/aapp/cmd"
	"github.com/ayachain/go-aya/chain"
	ChainCMD "github.com/ayachain/go-aya/chain/cmd"
	"github.com/ayachain/go-aya/keystore"
	KSCmd "github.com/ayachain/go-aya/keystore/cmd"
	ALogs "github.com/ayachain/go-aya/logs"
	TxCmd "github.com/ayachain/go-aya/tx/cmd"
	WalletCMD "github.com/ayachain/go-aya/wallet/cmd"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core"
	"sync"
)

var ayacmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Display file status.",
	},
	Subcommands: map[string]*cmds.Command{
		"aapp" 		: 	AappCMD.AAppCMDS,
		"keystore" 	: 	KSCmd.KeystoreCMDS,
		"chain"		: 	ChainCMD.ChainCMDS,
		"wallet"	:	WalletCMD.WalletCMDS,
		"tx"		:	TxCmd.TxCMDS,
	},
}

func DaemonAyaChain( ind *core.IpfsNode ) {

	ALogs.ConfigLogs()

	keystore.Init("/Users/apple/.aya/keystore", ind)

}

func ShutdownAyaChain() {

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {

		chain.DisconnectionAll()

		wg.Done()

	}()

	wg.Wait()
}