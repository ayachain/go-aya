package assets

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/ayachain/go-aya/vdb/common"
)

type Assets struct {
	common.RawDBCoder
	Version 	uint8
	Avail		uint64
	Vote		uint64
	Locked		uint64
}

func NewAssets( avail, vote, locked uint64 ) *Assets {

	return &Assets{
		Version:DRVer,
		Avail:avail,
		Vote:vote,
		Locked:locked,
	}

}

func (r *Assets) Encode() []byte {

	buf := bytes.NewBuffer([]byte{})

	buf.WriteByte( byte(r.Version) )
	buf.Write( common.BigEndianBytes(r.Avail) )
	buf.Write( common.BigEndianBytes(r.Vote) )
	buf.Write( common.BigEndianBytes(r.Locked) )

	return buf.Bytes()
}

func (r *Assets) Decode( bs []byte ) error {

	if len(bs) != 25 {
		return errors.New("incomplete data of assets")
	}

	rd := bytes.NewBuffer(bs)
	bfrd := bufio.NewReader( rd )

	avail 	:= make([]byte, 8)
	vote 	:= make([]byte, 8)
	lock 	:= make([]byte, 8)

	ver, _ := bfrd.ReadByte()
	_, _ = bfrd.Read(avail)
	_, _ = bfrd.Read(vote)
	_, _ = bfrd.Read(lock)

	r.Version = ver
	r.Avail = binary.BigEndian.Uint64(avail)
	r.Vote = binary.BigEndian.Uint64(vote)
	r.Locked = binary.BigEndian.Uint64(lock)

	return nil
}