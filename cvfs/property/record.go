package property

import (
	"bufio"
	"bytes"
	"encoding/json"
	"math/big"
)


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
	RecordEncoder
	RecordDecoder
	To			[]byte	`json:"t"`
	From		[]byte	`json:"f"`
	Count		[]byte	`json:"c"`
}

func (r *VotingRightRecord) DecodeKey( bs []byte ) error {

	brd := bytes.NewReader(bs)
	buf := bufio.NewReader(brd)

	to, err := buf.ReadSlice(',')
	if err != nil {
		return nil
	}

	_, err = buf.ReadByte()
	if err != nil {
		return err
	}

	from, _, err := buf.ReadLine()
	if err != nil {
		return nil
	}

	r.To = to
	r.From = from
	return nil
}

func (r *VotingRightRecord) EncodeKey() ([]byte, error) {

	buf := bytes.NewBuffer([]byte{})
	buf.Write( r.To )
	buf.WriteByte( '<' )
	buf.Write( r.From )

	return buf.Bytes(), nil
}

func (r *VotingRightRecord) Encode() ([]byte, error) {
	return r.Count, nil
}

func (r *VotingRightRecord) Decode( bs []byte ) error {
	r.Count = bs
	return nil
}
