package tx

import (
	"github.com/ayachain/go-aya/utils"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-ipfs-api"
)

type Block struct {
	Index	uint64
	Prev	string
	Txs[]	Tx
	BDHash	string		//Base Dir Ipfs Hash
	Hash	string		`json:"-"`
}

func NewBlock(index uint64, prev string, txs[] Tx, bdhash string) (b* Block){

	b = &Block{Index:index, Prev:prev, Txs:txs, BDHash:bdhash}

	return b
}

func (b *Block) GetHash() (bhash string, err error) {

	if len(b.Hash) > 0 {
		return b.Hash, nil
	}

	return b.WriteBlock()
}

func (b *Block) WriteBlock() (bhash string, err error) {

	bhash, err = utils.GetHash(b)

	if err != nil {
		return "", err
	}

	b.Hash = bhash

	return bhash, nil
}

func ReadBlock(hash string) (b *Block, err error) {

	bs, err := shell.NewLocalShell().BlockGet(hash)

	if err != nil {
		return nil, err
	}

	b = &Block{}
	if err :=json.Unmarshal(bs, b); err != nil {
		return nil, err
	}

	b.Hash = hash

	return b, nil
}

func (b *Block) PrintIndent() {

	bs, _ := json.MarshalIndent(b, "", "  ")

	fmt.Println(string(bs))
}