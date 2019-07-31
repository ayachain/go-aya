package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var poolCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "tx pool state command",
	},
	Subcommands: map[string]*cmds.Command{
		"state"	: stateCMD,
	},
}

var stateCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "tx pool state command",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}

		return ARsponse.EmitSuccessResponse(re, chain.GetTxPool().GetState())
	},
}