package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb/transaction"
	EComm "github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var publishCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "publish a signed hex encode raw transaction",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
		cmds.StringArg("signedrawtx", true, false, "signed tx raw hex encode string"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}

		rawTx := req.Arguments[1]

		txbs := EComm.Hex2Bytes(rawTx)
		if len(txbs) <= 0 {
			return ARsponse.EmitErrorResponse(re, errors.New("invalid transaction"))
		}

		tx := &transaction.Transaction{}
		if err := tx.Decode(txbs); err != nil {
			return ARsponse.EmitErrorResponse(re, errors.New("can not resolve transaction object"))
		}

		if tx.Verify() {
			return ARsponse.EmitErrorResponse(re, errors.New("transaction verify failed"))
		}

		c, err := chain.CVFServices().Transactions().GetTxCount(tx.From)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		if tx.Tid < c {
			return ARsponse.EmitErrorResponse(re, errors.New("transaction Tid is already in used"))
		}

		if err := chain.GetTxPool().PublishTx( tx ); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, tx.GetHash256().String())
	},
}