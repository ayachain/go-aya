package cmd

import (
	"context"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb/block"
	"github.com/ayachain/go-aya/vdb/indexes"
	"github.com/ayachain/go-aya/vdb/transaction"
	"github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

var getCmd = &cmds.Command {

	Helptext:cmds.HelpText{
		Tagline: "block detail command",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainID", true, false, "chain identifier"),
		cmds.StringArg("indexParams", false, false, "block index number or name ( LATEST, EARLIEST, MINING )"),
	},
	Options: []cmds.Option{
		cmds.BoolOption("idx", "i", "get detail"),
		cmds.BoolOption("txlist", "l", "get detail"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}

		blockParmas := "LATEST"

		if len(req.Arguments) > 1 {
			blockParmas = req.Arguments[1]
		}

		var (
			idx *indexes.Index
			err error
		)

		switch strings.ToUpper(blockParmas) {

		case "LATEST":
			idx, err = chain.CVFServices().Indexes().GetLatest()

		case "EARLIEST":
			idx, err = chain.CVFServices().Indexes().GetIndex(0)

		case "MINING":
			return ARsponse.EmitSuccessResponse(re, chain.GetTxPool().GetMiningBlock())

		default:

			var (
				bindex uint64
				cverr error
			)

			if len(blockParmas) > 2 && strings.EqualFold(strings.ToUpper(blockParmas[:2]),"0x") {
				bindex, cverr = strconv.ParseUint(blockParmas, 16, 64)
			} else {
				bindex, cverr = strconv.ParseUint(blockParmas, 10, 64)
			}

			if cverr != nil {
				return ARsponse.EmitErrorResponse(re, errors.New("Invalid block index or name"))
			}

			idx, err = chain.CVFServices().Indexes().GetIndex(bindex)
		}

		if err != nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not found block") )
		}

		getidx, _ := req.Options["idx"].(bool)
		txlist, _ := req.Options["txlist"].(bool)

		if getidx {

			return ARsponse.EmitSuccessResponse(re, idx)

		} else {

			blk, err := chain.CVFServices().Blocks().GetBlocks( idx.Hash )
			if err != nil {
				return ARsponse.EmitErrorResponse(re, err)
			}

			if len(blk) <= 0 {
				return ARsponse.EmitErrorResponse(re, errors.New("not found block"))
			}

			if txlist {

				ind, err := cmdenv.GetNode(env)
				if err != nil {
					return ARsponse.EmitErrorResponse(re, err)
				}

				rctx, rcancel := context.WithTimeout(req.Context, 25)
				defer rcancel()

				type BlockDetail struct {
					*block.Block
					TxList []*transaction.Transaction
				}

				bd := &BlockDetail{
					Block:blk[0],
					TxList:blk[0].ReadTxsFromDAG(rctx, ind),
				}

				return ARsponse.EmitSuccessResponse(re, bd)

			} else {

				return ARsponse.EmitSuccessResponse(re, blk[0])

			}
		}
	},
}
