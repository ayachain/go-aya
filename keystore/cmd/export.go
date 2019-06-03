package cmd

import (
	"encoding/json"
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ipfs/go-ipfs-cmds"
)

var exportCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Import keystore from node",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("address", true, false, "export account address"),
		cmds.StringArg("passphrase", true, false, "account unlock passphrase"),
		cmds.StringArg("kspassphrase", true, false, "export key store json content passphrase"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		aks := AKeyStore.ShareInstance()
		if aks == nil {
			return ARsponse.EmitErrorResponse(re, fmt.Errorf("keystore services instance error"))
		}

		acc, err := findAccount(req.Arguments[0])
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		keyJSon, err := AKeyStore.ShareInstance().Export(acc, req.Arguments[1], req.Arguments[2])
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		var rsmap map[string]interface{}
		err = json.Unmarshal([]byte(keyJSon), &rsmap)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, rsmap)
	},
}