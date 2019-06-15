package cmd

import (
	"errors"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"time"
)

var unlockCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "unlock wallet",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("passphrase", true, false, "account passphrase"),
	},
	Options: []cmds.Option{
		cmds.UintOption("time", "t", "unlock times by sec, default is 1 min"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		acc := AKeyStore.GetCoinBaseAddress()

		ulktime, _ := req.Options["time"].(uint)

		if ulktime <= 0 {
			ulktime = 60
		}

		if err := AKeyStore.ShareInstance().TimedUnlock(acc, req.Arguments[0], time.Duration(ulktime) * time.Second); err != nil {
			return ARsponse.EmitErrorResponse(re, err )
		}

		return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
	},
}

var lockCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "lock wallet",
	},
	Arguments: []cmds.Argument {

	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		acc := AKeyStore.GetCoinBaseAddress()

		if acc.Address.Big().Int64() <= 0 {
			return ARsponse.EmitErrorResponse(re, errors.New("coin base address not set") )
		}

		if err := AKeyStore.ShareInstance().Lock(acc.Address); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
	},
}