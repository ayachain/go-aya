package cmd

import (
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var listCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Show current keystore account lists.",
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		aks := AKeyStore.ShareInstance()
		if aks == nil {
			return ARsponse.EmitErrorResponse(re, fmt.Errorf("keystore services instance error"))
		}

		var adds []string
		for _, acc := range aks.Accounts() {
			adds = append(adds, acc.Address.Hex())
		}

		return ARsponse.EmitSuccessResponse(re, adds)
	},
}