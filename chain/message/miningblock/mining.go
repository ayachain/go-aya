package miningblock

import (
	"bytes"
	"encoding/gob"
	"errors"
	AMsgBlock "github.com/ayachain/go-aya/chain/message/block"
	"github.com/ipfs/go-cid"
)

const MessagePrefix = byte('m')

var (
	ErrMsgPrefix = errors.New("not a chain info message")
)


type MsgRawMiningBlock AMsgBlock.MsgRawBlock

func (msg *MsgRawMiningBlock) ConfirmByCid( cid cid.Cid ) *AMsgBlock.MsgRawBlock {

	blk := &AMsgBlock.MsgRawBlock{}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(msg); err != nil {
		return nil
	}

	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(blk); err != nil {
		return nil
	}

	blk.ExtraData = cid.String()

	return blk
}

func (msg *MsgRawMiningBlock) Prefix() byte{
	return MessagePrefix
}


func (msg *MsgRawMiningBlock) Encode() []byte {

	buff := bytes.NewBuffer([]byte{MessagePrefix})
	buff.Write( msg.Block.Encode() )

	return buff.Bytes()
}

func (msg *MsgRawMiningBlock) Decode(bs []byte) error {

	if bs[0] != MessagePrefix {
		return ErrMsgPrefix
	}

	return msg.Block.Decode(bs[1:])
}

