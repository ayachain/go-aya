package tx

import (
	"../../utils"
	"encoding/json"
	"github.com/ipfs/go-ipfs-api"
)

type Block struct {
	Index	uint64
	Prev	string
	Txs[]	Tx
	Hash	string		`json:"-"`
}

func NewBlock(index uint64, prev string, txs[] Tx) (b* Block){

	b = &Block{Index:index, Prev:prev, Txs:txs}

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

	return b, nil
}