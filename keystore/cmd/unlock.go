package cmd

import (
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ipfs/go-ipfs-cmds"
	"time"
)

const (
	unlockAccountTimeOptionKey = "time"
)

var unLockCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Unlock a wallet account",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("address", true, false, "account address"),
		cmds.StringArg("passphrase", true, false, "account passphrase"),
	},
	Options: []cmds.Option{
		cmds.UintOption(unlockAccountTimeOptionKey, "t", "unlock times by sec"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		addrHex := req.Arguments[0]
		pwd := req.Arguments[1]

		acc, err := findAccount(addrHex)
		if err != nil {
			return re.Emit(err)
		}

		ulktime, _ := req.Options[unlockAccountTimeOptionKey].(uint)
		if ulktime > 0 {
			if err := AKeyStore.ShareInstance().TimedUnlock(acc, pwd,  time.Duration(ulktime) * time.Second); err != nil {
				return re.Emit(err)
			}
		} else {
			if err := AKeyStore.ShareInstance().Unlock(acc, pwd); err != nil {
				return re.Emit(err)
			}
		}

		return re.Emit("Success.")
	},
	//PostRun:cmds.PostRunMap {
	//	cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
	//		return nil
	//	},
	//},
}