package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	EComm "github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var getCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "get transaction detail info",
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

		// find in txpool
		txhash := EComm.HexToHash(req.Arguments[1])
		if tx := chain.GetTxPool().GetTx(txhash); tx != nil {
			return ARsponse.EmitSuccessResponse(re, tx)
		}

		tx, err := chain.CVFServices().Transactions().GetTxByHash( EComm.HexToHash(req.Arguments[1]) )
		if err != nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist transaction hash") )
		}

		return ARsponse.EmitSuccessResponse(re, tx)
	},
}