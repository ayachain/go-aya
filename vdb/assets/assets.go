package assets

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-cid"
	"io/ioutil"
)

type Assets struct {
	common.RawDBCoder
	Version 	uint8
	Avail		uint64
	Vote		uint64
	ExtraCid 	cid.Cid
}

func (r *Assets) Encode() []byte {

	buf := bytes.NewBuffer([]byte{})

	buf.WriteByte( byte(r.Version) )
	buf.Write( common.BigEndianBytes(r.Avail) )
	buf.Write( common.BigEndianBytes(r.Vote) )
	buf.Write( r.ExtraCid.Bytes() )

	return buf.Bytes()
}

func (r *Assets) Decode( bs []byte ) error {

	rd := bytes.NewBuffer(bs)
	bfrd := bufio.NewReader( rd )

	avail := make([]byte, 8)
	vote := make([]byte, 8)

	ver, err := bfrd.ReadByte()
	if err != nil {
		return errors.New("decode expected by version")
	}

	_, err = bfrd.Read( avail )
	if err != nil {
		return errors.New("decode expected by avail")
	}

	_, err = bfrd.Read( vote )
	if err != nil {
		return errors.New("decode expected by vote")
	}

	cidbs, err := ioutil.ReadAll(bfrd)
	if err != nil {
		return errors.New("decode expected by extra cid")
	}

	ecid, err := cid.Cast(cidbs)
	if err != nil {
		return errors.New("decode expected by  cid")
	}

	r.Version = ver
	r.Avail = binary.BigEndian.Uint64(avail)
	r.Vote = binary.BigEndian.Uint64(vote)
	r.ExtraCid = ecid

	return nil
}