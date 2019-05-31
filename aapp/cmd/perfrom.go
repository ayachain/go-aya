package cmd

import (
	"fmt"
	"github.com/ayachain/go-aya/aapp"
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
			return re.Emit( fmt.Sprintf("%v is not a daemoned AAppServices", req.Arguments[0]) )
		}

		api := req.Arguments[1]

		var parmas []string

		for i := 2; i < len(req.Arguments); i++ {
			parmas = append(parmas, req.Arguments[i])
		}

		if ret, err := ap.Avm.PerfromGlobal(api, parmas...); err != nil {
			return re.Emit(err)
		} else {
			return re.Emit(ret)
		}

	},
}
