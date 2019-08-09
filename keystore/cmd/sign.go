package cmd

import (
	"encoding/hex"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-ipfs-cmds"
)

var signCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "sign a content with account",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("address", true, false, "account address"),
		cmds.StringArg("content", true, true, "sign content"),
		cmds.StringArg("passphrase", false, false, "account passphrase"),
	},
	Options: []cmds.Option{
		cmds.BoolOption("keccak256", "k", "content is Keccak256 hash string"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		var pwd = ""
		addrHex := req.Arguments[0]
		content := req.Arguments[1]

		if len(req.Arguments) > 2 {
			pwd = req.Arguments[2]
		}

		acc, err := findAccount(addrHex)

		if err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		var hash EComm.Hash

		contentIsKeccask256, _ := req.Options["keccak256"].(bool)
		if contentIsKeccask256 {

			hash = EComm.HexToHash( content )

		} else {

			hash = crypto.Keccak256Hash( []byte(content) )

		}

		if len(pwd) > 0 {

			signature, err := AKeyStore.ShareInstance().SignHashWithPassphrase(acc, pwd, hash.Bytes() )
			if err != nil {
				return ARsponse.EmitErrorResponse(re, err)
			} else {
				return ARsponse.EmitSuccessResponse(re, hex.EncodeToString(signature))
			}

		} else {

			signature, err := AKeyStore.ShareInstance().SignHash(acc, hash.Bytes())
			if err != nil {
				return ARsponse.EmitErrorResponse(re, err)
			} else {
				return ARsponse.EmitSuccessResponse(re, hex.EncodeToString(signature))
			}
		}
	},
}