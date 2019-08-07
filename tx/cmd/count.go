package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	EComm "github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var countCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "get address transaction total count in best chain",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
		cmds.StringArg("address", true, false, "owner address"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}

		c, err := chain.CVFServices().Transactions().GetTxCount( EComm.HexToAddress(req.Arguments[1]) )
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, c)
	},
}