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
	hash	string		`json:"-"`
}

func NewBlock(index uint64, prev string, txs[] Tx, bdhash string) (b* Block){

	b = &Block{Index:index, Prev:prev, Txs:txs, BDHash:bdhash}

	return b
}

func (b *Block) RefreshHash() string {

	hash, err := b.WriteBlock()

	if err != nil {
		return ""
	} else {
		b.hash = hash
		return hash
	}

}

func (b *Block) GetHash() (bhash string) {

	if len(b.hash) > 0 {
		return b.hash
	}

	hash, err := b.WriteBlock()

	if err != nil {
		return ""
	} else {
		return hash
	}
}

func (b *Block) WriteBlock() (bhash string, err error) {

	bhash, err = utils.GetHash(b)

	if err != nil {
		return "", err
	}

	b.hash = bhash

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

	b.hash = hash

	return b, nil
}

func (b *Block) PrintIndent() {

	bs, _ := json.MarshalIndent(b, "", "  ")

	fmt.Println(string(bs))
}