package vdb

import (
	"bytes"
	"github.com/ayachain/go-aya/vdb/assets"
	"github.com/ayachain/go-aya/vdb/block"
	"github.com/ayachain/go-aya/vdb/headers"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
	"time"
)

func TestACVFS_Assetses(t *testing.T) {

	east := &assets.Assets{
		Version:assets.DRVer,
		Avail:100000000,
		Vote:100000000,
		Locked:100000000,
	}
	encoding := east.Encode()

	oast := &assets.Assets{}
	if err := oast.Decode(encoding); err != nil {
		t.Fatalf("Assets RawDBCoder Test Failed")
	}

	if east.Version != oast.Version ||
		east.Avail != oast.Avail ||
		east.Vote != oast.Vote ||
		east.Locked != east.Locked {

		t.Fatalf("Assets RawDBCoder Test Failed")
	}

}


func TestACVFS_Blocks(t *testing.T) {

	eblock := &block.Block{
		Index:10000000,
		ChainID:"TestChainID",
		Parent:"532eaabd9574880dbf76b9b8cc00832c20a6ec113d682299550d7a6e0f345e25",
		ExtraData:"532eaabd9574880dbf76b9b8cc00832c20a6ec113d682299550d7a6e0f345e25",
		Timestamp:uint64(time.Now().Unix()),
		AppendData:[]byte("SomeAppendData"),
		Txc: 2,
		Txs: "532eaabd9574880dbf76b9b8cc00832c20a6ec113d682299550d7a6e0f345e25",
	}
	encoded := eblock.Encode()

	dblock := &block.Block{}
	if err := dblock.Decode(encoded); err != nil {
		t.Fatalf("Block RawDBCoder Test Failed")
	}

	if eblock.Index != dblock.Index ||
		eblock.ChainID != dblock.ChainID ||
		eblock.Parent != dblock.Parent ||
		eblock.ExtraData != dblock.ExtraData ||
		eblock.Timestamp != dblock.Timestamp ||
		!bytes.Equal(eblock.AppendData, dblock.AppendData) ||
		eblock.Txc != dblock.Txc ||
		eblock.Txs != dblock.Txs {
		t.Fatalf("Block RawDBCoder Test Failed")
	}


	encoded = eblock.RawMessageEncode()
	if err := dblock.RawMessageDecode(encoded); err != nil {
		t.Fatalf("Block AMessageCoder Test Failed")
	}

	if eblock.Index != dblock.Index ||
		eblock.ChainID != dblock.ChainID ||
		eblock.Parent != dblock.Parent ||
		eblock.ExtraData != dblock.ExtraData ||
		eblock.Timestamp != dblock.Timestamp ||
		!bytes.Equal(eblock.AppendData, dblock.AppendData) ||
		eblock.Txc != dblock.Txc ||
		eblock.Txs != dblock.Txs {
		t.Fatalf("Block AMessageCoder Test Failed")
	}

}


func TestACVFS_Headers(t *testing.T) {

	o := &headers.Header{
		BlockIndex:1000000,
		Hash:crypto.Keccak256Hash([]byte("Test")),
	}
	encoded := o.Encode()

	n := &headers.Header{}
	if err := n.Decode(encoded); err != nil {
		t.Fatalf("Header RawDBCoder Test Failed")
	}

	if o.BlockIndex != n.BlockIndex ||
		!bytes.Equal(o.Hash.Bytes(), n.Hash.Bytes()) {
		t.Fatalf("Header AMessageCoder Test Failed")
	}

}