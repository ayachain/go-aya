package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	EComm "github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var receiptCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "get transaction receipt message",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
		cmds.StringArg("txhash", true, false, "hex hash code by transaction."),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}

		tx, err := chain.CVFServices().Receipts().GetTransactionReceipt( EComm.HexToHash(req.Arguments[1]) )
		if err != nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist transaction hash or receipt") )
		}

		return ARsponse.EmitSuccessResponse(re, tx)
	},
}