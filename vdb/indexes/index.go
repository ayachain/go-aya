package indexes

import (
	"bytes"
	"encoding/binary"
	"github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
)

type Index struct {

	common.RawDBCoder		`json:"-"`

	BlockIndex 	uint64 		`json:"Index"`
	Hash		EComm.Hash 	`json:"Hash"`
	FullCID 	cid.Cid 	`json:"FullCID"`
}

const StaticSize = 74

func (i *Index) Encode() []byte {

	buff := bytes.NewBuffer([]byte{})

	buff.Write( common.BigEndianBytes(i.BlockIndex) )
	buff.Write( i.Hash.Bytes() )
	buff.Write( i.FullCID.Bytes() )

	return buff.Bytes()
}

func (i *Index) Decode(bs []byte) error {

	cid, err := cid.Cast( bs[8+32:] )
	if err != nil {
		return err
	}

	i.FullCID = cid
	i.BlockIndex = binary.BigEndian.Uint64( bs[:8] )
	i.Hash = EComm.BytesToHash( bs[8:8+32] )

	return nil
}