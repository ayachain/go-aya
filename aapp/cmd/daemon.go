package cmd

import (
	"fmt"
	"github.com/ayachain/go-aya/aapp"
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
			return re.Emit( err.Error() )
		}

		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return re.Emit( err.Error() )
		}

		_, err = aapp.Manager.Load(req.Arguments[0], api, nd)

		if err != nil {
			return re.Emit( err.Error() )
		}

		return re.Emit( fmt.Sprintf("Daemon AAPP : %v Success.", req.Arguments[0]) )
	},
	//PostRun:cmds.PostRunMap {
	//	cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
	//		return nil
	//	},
	//},
}