package cmd

import (
	AChain "github.com/ayachain/go-aya/chain"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	EAccount "github.com/ethereum/go-ethereum/accounts"
	"github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"github.com/pkg/errors"
	"io/ioutil"
	"strings"
)

var connCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "conn to a aya block chain by gen block json content",
	},
	Arguments: []cmds.Argument {
		cmds.FileArg("genblock", true, false, "first block json content"),
		cmds.StringArg("authaddress", true, false, "default user address"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		ind, err := cmdenv.GetNode(env)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, errors.New("ipfs node get failed"))
		}

		file, err := cmdenv.GetFileArg(req.Files.Entries())
		if err != nil {
			return ARsponse.EmitSuccessResponse(re, err)
		}

		bs, err := ioutil.ReadAll(file)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		gblk := &ABlock.GenBlock{}
		if err := gblk.Decode(bs); err != nil {
			return ARsponse.EmitErrorResponse(re, errors.New("decode gen block config file expected") )
		}

		acc, err := AKeyStore.FindAccount( req.Arguments[0] )
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err )
		}

		if err := AChain.AddChainLink(gblk, ind, acc); err != nil {
			return ARsponse.EmitErrorResponse(re, err )
		}

		return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
	},
}




func findAccount( hexAddr string ) (*EAccount.Account, error) {

	aks := AKeyStore.ShareInstance()
	if aks == nil {
		return nil, errors.New( "AKeyStore services expected" )
	}

	for _, acc := range aks.Accounts() {
		if strings.EqualFold( acc.Address.String(), hexAddr ) {
			return &acc, nil
		}
	}

	return nil, errors.New("address not found")
}