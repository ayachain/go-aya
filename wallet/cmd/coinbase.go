package cmd

import (
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var coinBaseGet = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "get coin base address from wallet",
	},
	Arguments: []cmds.Argument {

	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		acc := AKeyStore.GetCoinBaseAddress()

		return ARsponse.EmitSuccessResponse(re, acc.Address.String())
	},
}

var coinBaseSet = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "set coin base address to wallet",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("address", true, false, "accounts address"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		acc, err := AKeyStore.FindAccount(req.Arguments[0])
		if err != nil {
			return ARsponse.EmitErrorResponse(re, fmt.Errorf("%v not found address", req.Arguments[0]))
		}

		if err := AKeyStore.SetCoinBaseAddress(acc); err != nil {
			return ARsponse.EmitSuccessResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
	},
}


var coinBaseCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "set coin base address from wallet",
	},
	Subcommands: map[string]*cmds.Command{
		"get" : coinBaseGet,
		"set" : coinBaseSet,
	},
}