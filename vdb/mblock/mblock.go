package mblock

import (
	"bytes"
	"encoding/gob"
	"errors"
	ABlock "github.com/ayachain/go-aya/vdb/block"
)

const MessagePrefix = byte('m')

var (
	ErrMsgPrefix = errors.New("not a chain info message")
)

type MBlock struct {
	ABlock.Block
}

func (mb *MBlock) Confirm( groupCid string ) *ABlock.Block {

	blk := &ABlock.Block{}
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(mb.Block); err != nil {
		return nil
	}

	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(blk); err != nil {
		return nil
	}

	blk.ExtraData = groupCid

	return blk

}


func ( b *MBlock ) RawMessageEncode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})

	buff.Write( b.Encode() )

	return buff.Bytes()
}

func ( b *MBlock ) RawMessageDecode( bs []byte ) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return b.Decode(bs[1:])

}