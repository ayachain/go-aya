package cmd

import (
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
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
			return ARsponse.EmitErrorResponse(re, fmt.Errorf("keystore services instance error"))
		}

		pwd := req.Arguments[0]
		acc, err := aks.NewAccount(pwd)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, acc.Address.Hex() )
	},
}