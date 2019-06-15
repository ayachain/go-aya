package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	ATxUtils "github.com/ayachain/go-aya/tx/txutils"
	EComm "github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"strconv"
)


var transaferAvail = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "transfer balance",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
		cmds.StringArg("to", true, false, "target address"),
		cmds.StringArg("value", true, false, "amount"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}


		acc := AKeyStore.GetCoinBaseAddress()

		vnumber, err := strconv.ParseUint(req.Arguments[2], 10, 64)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		txcount, err := chain.CVFServices().Transactions().GetTxCount(acc.Address)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		tx := ATxUtils.MakeTransferAvail(acc.Address, EComm.HexToAddress(req.Arguments[1]), vnumber, txcount + 1)

		if err := AKeyStore.SignTransaction(tx, acc); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		if err := chain.SendRawMessage(tx); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, tx.GetHash256())
	},

}


var transaferDoLock = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "lock avail balance",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
		cmds.StringArg("value", true, false, "amount"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}


		acc := AKeyStore.GetCoinBaseAddress()

		vnumber, err := strconv.ParseUint(req.Arguments[2], 10, 64)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		txcount, err := chain.CVFServices().Transactions().GetTxCount(acc.Address)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		tx := ATxUtils.MakeTransferAvail(acc.Address, EComm.HexToAddress(req.Arguments[1]), vnumber, txcount)

		if err := AKeyStore.SignTransaction(tx, acc); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		if err := chain.SendRawMessage(tx); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, tx.GetHash256())
	},
}


var transferCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "transfer balance",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
		cmds.StringArg("to", true, false, "target address"),
		cmds.StringArg("value", true, false, "amount"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}


		acc := AKeyStore.GetCoinBaseAddress()

		vnumber, err := strconv.ParseUint(req.Arguments[2], 10, 64)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		txcount, err := chain.CVFServices().Transactions().GetTxCount(acc.Address)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		tx := ATxUtils.MakeTransferAvail(acc.Address, EComm.HexToAddress(req.Arguments[1]), vnumber, txcount)

		if err := AKeyStore.SignTransaction(tx, acc); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		if err := chain.SendRawMessage(tx); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, tx.GetHash256())
	},
}