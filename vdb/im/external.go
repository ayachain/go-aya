package im

import (
	"bytes"
	"context"
	"encoding/json"
	VDBComm "github.com/ayachain/go-aya/vdb/common"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
)

func ( b *Block ) GetHash() EComm.Hash {

	bs, err := proto.Marshal(b)
	if err != nil {
		panic(err)
	}

	return crypto.Keccak256Hash(bs)
}

func ( m *Block ) ReadTxsFromDAG(ctx context.Context, ind *core.IpfsNode) []*Transaction {

	dcid, err := cid.Cast(m.Txs)
	if err != nil {
		return nil
	}

	iblk, err := ind.Blocks.GetBlock(ctx, dcid)
	if err != nil {
		return nil
	}

	txlist := make([]*Transaction, m.Txc)

	if err := json.Unmarshal( iblk.RawData(), &txlist ); err != nil {
		return nil
	} else {
		return txlist
	}

}

func ( m *GenBlock ) GetHash() EComm.Hash {

	bs, err := proto.Marshal(m)
	if err != nil {
		panic("unrecoverable computing exception : Hash")
	}

	if bs == nil {
		panic("unrecoverable computing exception : Hash")
	}

	return crypto.Keccak256Hash(bs)
}

func ( m *ChainInfo ) GetHash() EComm.Hash {

	bs, err := proto.Marshal(m)
	if err != nil {
		panic("unrecoverable computing exception : Hash")
	}

	if bs == nil {
		panic("unrecoverable computing exception : Hash")
	}

	return crypto.Keccak256Hash(bs)
}

func ( m *Transaction ) GetHash256( ) EComm.Hash {

	buff := bytes.NewBuffer([]byte("AyaTransactionPrefix"))

	buff.Write( VDBComm.BigEndianBytes(m.BlockIndex) )
	buff.Write( m.From )
	buff.Write( m.To )
	buff.Write( VDBComm.BigEndianBytes(m.Value) )
	buff.Write( []byte(m.Data) )
	buff.Write( VDBComm.BigEndianBytes(m.Tid) )
	buff.Write( VDBComm.BigEndianBytesUint16(uint16(m.Type)))
	//buff.Write( trsn.Sig )

	return crypto.Keccak256Hash( buff.Bytes() )
}

func ( m *Transaction ) Verify() bool {

	hs := m.GetHash256()

	pubkey, err := crypto.SigToPub(hs.Bytes(),m.Sig )
	if err != nil {
		return false
	}

	from := crypto.PubkeyToAddress(*pubkey)

	return bytes.Equal(from.Bytes(), m.From)
}

func ConfirmReceipt( msg string, events []string ) *Receipt {

	return &Receipt{
		Stat:0,
		Message:msg,
		Event:events,
	}

}

func ExpectedReceipt( msg string, events []string ) *Receipt {

	return &Receipt{
		Stat:-1,
		Message:msg,
		Event:events,
	}
}