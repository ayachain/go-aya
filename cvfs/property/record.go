package property

import (
	"errors"
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/ipfs/go-cid"
	"io/ioutil"
	"math/big"
)

type Record struct {
	Version 	uint8
	Avail			uint64
	Vote			uint64
	ExtraCid 	cid.Cid
}

func bigEndianBytes (number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func (r *Record) Encode( ) []byte {

	buf := bytes.NewBuffer([]byte{})

	buf.WriteByte( byte(r.Version) )
	buf.Write( bigEndianBytes(r.Avail) )
	buf.Write( bigEndianBytes(r.Vote) )
	buf.Write( r.ExtraCid.Bytes() )

	return buf.Bytes()
}

func (r *Record) Decode( bs []byte ) error {

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

type RecordEncoder interface {
	Encode() ([]byte, error)
}

type RecordDecoder interface {
	Decode( []byte ) error
}

type AvailRecord struct {
	*big.Int
	RecordEncoder
	RecordDecoder
}

func (r *AvailRecord) Encode() ([]byte, error) {
	return r.Int.GobEncode()
}

func (r *AvailRecord) Decode( bs []byte ) error {
	return r.Int.GobDecode(bs)
}

type FrozenRecord struct {
	RecordEncoder
	RecordDecoder
	Total 		[]byte	`json:"t"`
	Start 		[]byte	`json:"s"`
	End   		[]byte	`json:"e"`
	Released	[]byte	`json:"r"`
}

func (r *FrozenRecord) Encode() ([]byte, error) {
	return json.Marshal(r)
}

func (r *FrozenRecord) Decode( bs []byte ) error {
	return json.Unmarshal(bs, r)
}

type VotingRightRecord struct {
	*big.Int
	RecordEncoder
	RecordDecoder
}

func (r *VotingRightRecord) Encode() ([]byte, error) {
	return r.Int.GobEncode()
}

func (r *VotingRightRecord) Decode( bs []byte ) error {
	return r.Int.GobDecode(bs)
}
