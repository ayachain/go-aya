package cmd

import (
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ipfs/go-ipfs-cmds"
)

var deleteAccountCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "delete a account",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("address", true, false, "accounts address"),
		cmds.StringArg("passphrase", true, false, "accounts passphrase"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		aks := AKeyStore.ShareInstance()
		if aks == nil {
			return re.Emit("keystore services instance error")
		}

		addrHex := req.Arguments[0]
		acc, err := findAccount(addrHex)
		if err != nil {
			return re.Emit(err)
		}

		pwd := req.Arguments[1]
		if err := aks.Delete(acc, pwd); err != nil {
			return re.Emit("delete account success")
		} else {
			return re.Emit(err)
		}
	},
}