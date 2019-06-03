package cmd

import (
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"io/ioutil"
)

var importCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Import keystore from node",
	},
	Arguments: []cmds.Argument {
		cmds.FileArg("keystore", true, false, "keystore json content file"),
		cmds.StringArg("passphrase", true, false, "keystore passphrase"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		aks := AKeyStore.ShareInstance()
		if aks == nil {
			return ARsponse.EmitErrorResponse(re, fmt.Errorf("keystore services instance error"))
		}


		file, err := cmdenv.GetFileArg(req.Files.Entries())
		if err != nil {
			return ARsponse.EmitSuccessResponse(re, err)
		}
		//defer file.Close()

		ksjson, err := ioutil.ReadAll(file)
		if err != nil {
			return ARsponse.EmitSuccessResponse(re, err)
		}

		acc, err := AKeyStore.ShareInstance().Import(ksjson, req.Arguments[0], req.Arguments[0])
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		} else {
			return ARsponse.EmitSuccessResponse(re, acc.Address.String())
		}

	},
}