package tx

import (
	"github.com/ayachain/go-aya/utils"
	"container/list"
	"github.com/pkg/errors"
	"strings"
)

type chain struct {
	LatestBlock *Block
	LatestHash  string
}

func (c *chain) LinkChaih(latestHash string) (err error){

	b, err := ReadBlock(latestHash)

	if err != nil {
		return err
	}

	c.LatestHash = latestHash
	c.LatestBlock = b

	return nil
}

func (c *chain) GenChain(gb *Block) (err error) {

	hstr, err := utils.GetHash(gb)

	if err != nil {
		return err
	}

	c.LatestHash = hstr
	c.LatestBlock = gb

	return nil
}

func (c *chain) AppendBlock(b *Block) (err error) {

	if strings.EqualFold(b.Prev, c.LatestHash) {

		bhash, err := b.WriteBlock()

		if err != nil {
			return err
		}

		c.LatestBlock = b
		c.LatestHash = bhash

		return nil
	}

	return errors.New("New block record prev block not equal latest hashcode.")

}


type memChain struct {
	chain
	blocks list.List
}


func (c *memChain) LinkChaih(latestHash string) (err error){

	if err := c.chain.LinkChaih(latestHash); err != nil {
		return err
	}

	c.blocks.PushBack(c.LatestBlock)

	return nil
}

func (c *memChain) GenChain(gb *Block) (err error) {

	if err := c.chain.GenChain(gb); err != nil {
		return err
	}

	c.blocks.PushBack(gb)

	return nil
}

func (c *memChain) AppendBlock(b *Block) (err error) {

	if err := c.chain.AppendBlock(b); err != nil {
		return err
	}

	c.blocks.PushBack(b)

	return nil
}


func (c *memChain) SearchBlockByPrev(prevHash string) (r* Block, err error) {

	it := c.blocks.Front()

	for it != nil {

		if strings.EqualFold(it.Value.(*Block).Prev, prevHash) {
			return it.Value.(*Block), nil
		}

		it = it.Next()
	}

	return nil, errors.New("NotFound block in memory chain.")

}