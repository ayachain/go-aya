package block

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
)

const (
	Genesis = 0
	Latest  = ^uint64(0)
)

func BlockEqual( a *Block, b *Block ) bool {

	abs, _ := json.Marshal(a)
	bbs, _ := json.Marshal(b)

	return bytes.Equal(abs, bbs)
}

func GetBlocks( ind *core.IpfsNode, c ... cid.Cid ) ([]*Block, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var blks []*Block

	cchan := ind.Blocks.GetBlocks( ctx, c )

	for {

		b, closed := <- cchan

		if closed {
			break
		}

		sblk := &Block{}

		if err := json.Unmarshal(b.RawData(), sblk); err != nil {
			return nil, err
		}

		blks = append(blks, sblk)
	}

	return blks, nil
}