package cmd

import (
	"github.com/ayachain/go-aya/aapp"
	ARsponse "github.com/ayachain/go-aya/response"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var listCmd = &cmds.Command{
	Helptext:cmds.HelpText{
		Tagline: "Show all daemoned aapps.",
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		list := aapp.Manager.List()

		return ARsponse.EmitSuccessResponse(re, list)

	},
}