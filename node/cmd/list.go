package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	ANode "github.com/ayachain/go-aya/vdb/node"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/syndtr/goleveldb/leveldb"
)

var listCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "get aya network node identifier list",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])

		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}

		var retlist[] *ANode.Node
		if err := chain.CVFServices().Nodes().DoRead(func(db *leveldb.DB) error {

			it := db.NewIterator(nil, nil)
			defer it.Release()

			for it.Next() {

				nd := &ANode.Node{}

				if err := nd.Decode(it.Value()); err != nil {
					continue
				} else {
					retlist = append(retlist, nd)
				}
			}

			return nil

		}); err != nil {
			return ARsponse.EmitErrorResponse(re, err)
		}

		return ARsponse.EmitSuccessResponse(re, retlist)
	},
}