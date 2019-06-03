package cmd

import (
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ipfs/go-ipfs-cmds"
)

var updateCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "modify wallet keystore passphrase",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("address", true, false, "account address"),
		cmds.StringArg("old", true, false, "origin account passphrase"),
		cmds.StringArg("new", true, false, "new account passphrase"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		addrHex := req.Arguments[0]
		oldPwd := req.Arguments[1]
		newPwd := req.Arguments[2]

		acc, err := findAccount(addrHex)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		if err := AKeyStore.ShareInstance().Update(acc, oldPwd, newPwd); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
	},
}