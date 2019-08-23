package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb/im"
	EComm "github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/syndtr/goleveldb/leveldb"
)


var balanceOfCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "get address asset detail info",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
		cmds.StringArg("address", false, false, "target address default is coinbase address"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}

		var addr EComm.Address
		if len(req.Arguments) > 1 {

			addr = EComm.HexToAddress(req.Arguments[1])
		} else {

			addr = AKeyStore.GetCoinBaseAddress().Address
		}

		ast, err := chain.CVFServices().Assetses().AssetsOf(addr)

		if err != nil {

			if err == leveldb.ErrNotFound {
				return ARsponse.EmitSuccessResponse(re, &im.Assets{ Avail:0, Vote:0, Locked:0, })
			} else {
				return ARsponse.EmitErrorResponse(re, err)
			}

		}

		return ARsponse.EmitSuccessResponse(re, ast)
	},
}