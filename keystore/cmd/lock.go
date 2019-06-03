package cmd

import (
	"fmt"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-ipfs-cmds"
)

const lockAccountAllOpetionKey  = "all"

var lockCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "Unlock a wallet account",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("address", false, false, "lock account address"),
	},
	Options: []cmds.Option{
		cmds.BoolOption(lockAccountAllOpetionKey, "a", "lock all unlocked accounts"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		var addrHex string

		if len(req.Arguments) > 0 {
			addrHex = req.Arguments[0]
		}

		lockAll := req.Options[lockAccountAllOpetionKey].(bool)


		if lockAll {

			for _, acc := range AKeyStore.ShareInstance().Accounts() {

				if err := AKeyStore.ShareInstance().Lock(acc.Address); err != nil {

					return ARsponse.EmitErrorResponse(re, err)

				}

			}

			return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)

		} else if len(addrHex) > 0 {

			if err := AKeyStore.ShareInstance().Lock( common.HexToAddress(addrHex) ); err != nil {

				return ARsponse.EmitErrorResponse(re, err)

			} else {

				return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)

			}

		} else {

			return ARsponse.EmitErrorResponse(re, fmt.Errorf("miss account address params"))

		}

	},
}