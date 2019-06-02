package cmd

import (
	AKeyStore "github.com/ayachain/go-aya/keystore"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var accountsCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Show current keystore account lists.",
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		aks := AKeyStore.ShareInstance()
		if aks == nil {
			return re.Emit("keystore services instance error")
		}

		var adds []string
		for _, acc := range aks.Accounts() {
			adds = append(adds, acc.Address.Hex())
		}

		return re.Emit(adds)
	},
}