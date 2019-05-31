package cmd

import (
	"github.com/ayachain/go-aya/aapp"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var listCmd = &cmds.Command{
	Helptext:cmds.HelpText{
		Tagline: "Show all daemoned aapps.",
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		list := aapp.Manager.List()
		return re.Emit(list)
	},
}