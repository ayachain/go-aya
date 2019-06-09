package cmd

import (
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	cmds "github.com/ipfs/go-ipfs-cmds"
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