package cmd

import (
	"fmt"
	"github.com/ayachain/go-aya/aapp"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var shutdownCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "shutdown a daemoned AApp",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("aappns", true, false, "Path to AApp."),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		if err := aapp.Manager.Shutdown(req.Arguments[0]); err != nil {
			return err
		} else {
			return re.Emit( fmt.Sprintf("Shutdown AAPP : %v Success.", req.Arguments[0]) )
		}
	},
}