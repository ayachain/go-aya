package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/golang/protobuf/proto"
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

		var retlist[] *im.Node
		if err := chain.CVFServices().Nodes().DoRead(func(db *leveldb.DB) error {

			it := db.NewIterator(nil, nil)
			defer it.Release()

			for it.Next() {

				nd := &im.Node{}
				if err := proto.Unmarshal(it.Value(), nd); err != nil {
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