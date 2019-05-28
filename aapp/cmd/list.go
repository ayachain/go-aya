package cmd

import (
	"encoding/json"
	"github.com/ayachain/go-aya/aapp"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var listCmd = &cmds.Command{
	Helptext:cmds.HelpText{
		Tagline: "Show all daemoned aapps.",
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		list := aapp.Manager.List()

		if bs, err := json.Marshal(list); err != nil {
			return re.Emit(err.Error())
		} else {
			return re.Emit(string(bs))
		}

	},
}