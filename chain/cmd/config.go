package cmd

import (
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"github.com/pkg/errors"
)

var configCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "init aya chain connection",
	},
	Options: []cmds.Option{
		cmds.BoolOption("reload", "r", "if direct substitution already exists"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		ind, err := cmdenv.GetNode(env)
		if err != nil {
			return ARsponse.EmitErrorResponse(re, errors.New("ipfs node get failed"))
		}

		gblk := &im.GenBlock{
			ChainID:"main",
			Parent:"f919678e3089f1da5b4eac138c86efa0461555223f71f0d53a5a1ea9472834a0",
			Index:0,
			Timestamp:1564199977,
			Txc:0,
			Txs:cid.Undef.Bytes(),
			ExtraData:nil,
			AppendData:[]byte("SSUyMHJlc2VydmVkJTIwYW4lMjBpbXBvcnRhbnQlMjBtZXNzYWdlJTIwaW4lMjBwYXJlbnQlMjBoYXNoLg=="),
			Award: []*im.GenAssets {
				{
					Avail:1000000000000,
					Vote:1000000000000,
					Locked:0,
					Owner:common.HexToAddress("0xB137D22eD06f22A3C6E2C5fdeAB49C612678bd6F").Bytes(),
				},
				{
					Avail:1000000000000,
					Vote:1000000000000,
					Locked:0,
					Owner:common.HexToAddress("0x17C6594CCf4798DB4eF9fE169A8C269847520D3c").Bytes(),
				},
				{
					Avail:1000000000000,
					Vote:1000000000000,
					Locked:0,
					Owner:common.HexToAddress("0x8E0A21BCDb37eF545f4b04ef30a773F137DCb372").Bytes(),
				},
			},
			SuperNodes: []*im.Node{
				{
					Type:im.NodeType_Super,
					Votes:1000000000000,
					PeerID:"QmX9s1TVMiP3u9k4ZRvqhu3fWXnVF6Q5hsBHV2oRwNsTka",
					Owner:common.HexToAddress("0xB137D22eD06f22A3C6E2C5fdeAB49C612678bd6F").Bytes(),
					Sig:[]byte(""),
				},
				{
					Type:im.NodeType_Super,
					Votes:1000000000000,
					PeerID:"QmRkgBZcs8sSFW6ruksT9jvSjXgq29bTVguMW2zaUX7nrx",
					Owner:common.HexToAddress("0x17C6594CCf4798DB4eF9fE169A8C269847520D3c").Bytes(),
					Sig:[]byte(""),
				},
				{
					Type:im.NodeType_Super,
					Votes:1000000000000,
					PeerID:"QmQgTSEAq2Q5adZrm8NaQ1dkjhQhbwoQ4d5v6kKgoRcbdr",
					Owner:common.HexToAddress("0x8E0A21BCDb37eF545f4b04ef30a773F137DCb372").Bytes(),
					Sig:[]byte(""),
				},
			},
		}

		r, _ := req.Options["reload"].(bool)

		if err := AChain.AddChain(gblk, ind, r); err != nil {
			return ARsponse.EmitErrorResponse(re, err )
		}

		return ARsponse.EmitSuccessResponse(re, ARsponse.SimpleSuccessBody)
	},
}