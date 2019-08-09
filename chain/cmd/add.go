package cmd

import (
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"github.com/pkg/errors"
	"io/ioutil"
)

var addCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "init aya chain connection",
	},
	Arguments: []cmds.Argument {
		cmds.FileArg("genblock", true, false, "first block json content"),
	},
	Options: []cmds.Option{
		cmds.BoolOption("replace", "r", "if direct substitution already exists"),
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

		r, _ := req.Options["replace"].(bool)

		if err := AChain.AddChain(gblk, ind, r); err != nil {
			return ARsponse.EmitErrorResponse(re, err )
		}

		return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
	},
}