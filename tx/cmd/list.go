package cmd

import (
	"errors"
	AChain "github.com/ayachain/go-aya/chain"
	ARsponse "github.com/ayachain/go-aya/response"
	"github.com/ayachain/go-aya/vdb/transaction"
	EComm "github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

var listCMD = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "get transaction list or hash",
	},
	Arguments: []cmds.Argument {
		cmds.StringArg("chainid", true, false, "aya chain id"),
		cmds.StringArg("address", true, false, "owner address"),
	},
	Options: []cmds.Option{
		cmds.BoolOption("detail", "d", "response tx content detail (default:false)"),
		cmds.UintOption("offset", "o", "transaction offset (default:0)"),
		cmds.UintOption("size", "s", "transaction size (default:20)"),
	},
	Run:func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {

		chain := AChain.GetChainByIdentifier(req.Arguments[0])
		if chain == nil {
			return ARsponse.EmitErrorResponse(re, errors.New("not exist chain connection") )
		}

		detail, _ := req.Options["detail"].(bool)
		offset, _ := req.Options["offset"].(uint)
		size, _ := req.Options["size"].(uint)

		if size == 0 {
			size = 20
		}

		if !detail {

			txhashs := chain.CVFServices().Transactions().GetHistoryHash( EComm.HexToAddress(req.Arguments[1]), uint64(offset), uint64(size) )

			return ARsponse.EmitSuccessResponse(re, txhashs)

		} else {

			txs, err := chain.CVFServices().Transactions().GetHistoryContent( EComm.HexToAddress(req.Arguments[1]), uint64(offset), uint64(size) )

			if err != nil {

				return ARsponse.EmitErrorResponse(re, errors.New("not exist transaction hash") )

			} else {

				var ctxs []*transaction.ConfirmTx

				for _, tx := range txs {

					txblock, err := chain.CVFServices().Blocks().GetBlocks(tx.BlockIndex)
					if err != nil {
						continue
					}

					ctxs = append(ctxs, &transaction.ConfirmTx{
						Transaction:*tx,
						Time:txblock[0].Timestamp,
					})

				}

				return ARsponse.EmitSuccessResponse(re, txs)

			}

		}
	},
}