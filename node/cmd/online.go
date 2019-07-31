package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	"github.com/ayachain/go-aya/chain/txpool"
	ARsponse "github.com/ayachain/go-aya/response"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var cstatCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "get aya network node online SuperNode state and ping RTT",
	},
	Arguments: []cmds.Argument {

		cmds.StringArg("chainid", true, false, "aya chain id"),

	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])

		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}

		if chain.GetTxPool().GetWorkMode() != txpool.AtxPoolWorkModeSuper {
			return ARsponse.EmitErrorResponse(re, errors.New("this method only valid in SuperNode work mode") )
		}

		return ARsponse.EmitSuccessResponse(re, chain.GetTxPool().ElectoralServices().GetNodesPingStates())
	},
}