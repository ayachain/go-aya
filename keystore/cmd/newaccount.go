package cmd

import (
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ipfs/go-ipfs-cmds"
)

var newAccountCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Create new account",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("passphrase", true, false, "accounts passphrase"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		aks := AKeyStore.ShareInstance()
		if aks == nil {
			return re.Emit("keystore services instance error")
		}

		pwd := req.Arguments[0]
		acc, err := aks.NewAccount(pwd)
		if err != nil {
			return re.Emit(err)
		}

		return re.Emit(acc.Address.Hex())
	},
}