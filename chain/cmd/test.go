package cmd

import (
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"strconv"
)

var testCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "do test",
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		err := AChain.GetChainByIdentifier("aya").Test()

		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		} else {
			return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
		}


	},
}


var indexOfCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "do test",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("index", true, false, "default user address"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		indexstr := req.Arguments[0]

		index, err := strconv.ParseUint(indexstr, 10, 64)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		ix, err := AChain.GetChainByIdentifier("aya").IndexOf( index )
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, ix)

	},
}