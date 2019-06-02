package cmd

import (
	"encoding/hex"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-ipfs-cmds"
)

var signCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "sign a content with account",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("address", true, false, "account address"),
		cmds.StringArg("content", true, false, "sign content"),
		cmds.StringArg("passphrase", true, false, "account passphrase"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		addrHex := req.Arguments[0]
		content := req.Arguments[1]
		pwd := req.Arguments[2]

		acc, err := findAccount(addrHex)
		if err != nil {
			return re.Emit(err)
		}

		hash := crypto.Keccak256Hash( []byte(content) )

		if len(pwd) > 0 {

			signature, err := AKeyStore.ShareInstance().SignHashWithPassphrase(acc, pwd, hash.Bytes() )
			if err != nil {
				return re.Emit(err)
			} else {
				return re.Emit(hex.EncodeToString(signature))
			}

		} else {

			signature, err := AKeyStore.ShareInstance().SignHash(acc, hash.Bytes())
			if err != nil {
				return re.Emit(err)
			} else {
				return re.Emit(hex.EncodeToString(signature))
			}
		}
	},
}