package block

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
)

const MessagePrefix = byte('b')

const (
	Genesis = 0
	Latest  = ^uint64(0)

	BlockNameLatest = "latest"
	BlockNameGen 	= "genesis"
)

func GetBlocks( ind *core.IpfsNode, c ... cid.Cid ) ([]*im.Block, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var blks []*im.Block

	cchan := ind.Blocks.GetBlocks( ctx, c )

	for {

		b, closed := <- cchan

		if closed {
			break
		}

		sblk := &im.Block{}

		if err := json.Unmarshal(b.RawData(), sblk); err != nil {
			return nil, err
		}

		blks = append(blks, sblk)
	}

	return blks, nil
}

func ConfirmBlock( block *im.Block, extCid cid.Cid ) *im.Block {

	blk := &im.Block{}
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(block); err != nil {
		return nil
	}

	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(blk); err != nil {
		return nil
	}

	blk.ExtraData = extCid.Bytes()

	return blk
}