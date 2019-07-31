package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	ANode "github.com/ayachain/go-aya/vdb/node"
	cmds "github.com/ipfs/go-ipfs-cmds"
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

		it := chain.CVFServices().Nodes().GetSnapshot().NewIterator(nil, nil)

		var retlist[] *ANode.Node

		for it.Next() {

			nd := &ANode.Node{}

			if err := nd.Decode(it.Value()); err != nil {
				continue
			} else {
				retlist = append(retlist, nd)
			}
		}

		return ARsponse.EmitSuccessResponse(re, retlist)
	},
}