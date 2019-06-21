package cmd

import (
	AChain "github.com/ayachain/go-aya/chain"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"github.com/pkg/errors"
	"io/ioutil"
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

		if err := AChain.AddChainLink(req.Context, gblk, ind, acc); err != nil {
			return ARsponse.EmitErrorResponse(re, err )
		}

		return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
	},
}


var dissConnectCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "dissconnect chain",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chainid"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		cc := AChain.GetChainByIdentifier( req.Arguments[0] )

		if cc == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("chain not connection") )
		}

		cc.Disconnect()

		return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
	},

}