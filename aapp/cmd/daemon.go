package cmd

import (
	"github.com/ayachain/go-aya/aapp"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
)

var daemonCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Start a new AApp from aappns path.",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("aappns", true, false, "Path to AApp."),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		_, err = aapp.Manager.Load(req.Arguments[0], api, nd)

		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(
			re,
			ARsponse.SimpleSuccessBody,
			)
	},
}