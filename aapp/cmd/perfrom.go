package cmd

import (
	"fmt"
	"github.com/ayachain/go-aya/aapp"
	ARsponse "github.com/ayachain/go-aya/response"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var perfromCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "perfrom a AApp method api",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("aappns", true, false, "Path to AApp."),
		cmds.StringArg("api", true, false, "Api name"),
		cmds.StringArg("args", false, true, "Api parmas"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		ap := aapp.Manager.AAppOf(req.Arguments[0])

		if ap == nil {
			return ARsponse.EmitErrorResponse(
				re,
				fmt.Errorf("%v is not a daemoned AAppServices", req.Arguments[0]),
			)
		}

		api := req.Arguments[1]

		var params []string

		for i := 2; i < len(req.Arguments); i++ {
			params = append(params, req.Arguments[i])
		}

		if ret, err := ap.Avm.PerfromGlobal(api, params...); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		} else {
			return ARsponse.EmitSuccessResponse(re, ret)
		}

	},
}
