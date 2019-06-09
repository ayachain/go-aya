package mblock

import (
	"bytes"
	"encoding/gob"
	"errors"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ipfs/go-cid"
)

const MessagePrefix = byte('m')

var (
	ErrMsgPrefix = errors.New("not a chain info message")
)

type MBlock struct {
	ABlock.Block
}

func (mb *MBlock) Confirm( bindex uint64,  cid cid.Cid ) *ABlock.Block {

	blk := &ABlock.Block{}
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(mb); err != nil {
		return nil
	}

	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(blk); err != nil {
		return nil
	}

	blk.ExtraData = cid.String()
	blk.Index = bindex

	return blk

}