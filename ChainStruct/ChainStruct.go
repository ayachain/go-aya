package ChainStruct

import (
	"container/list"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"log"
	"strconv"
	"time"
)

type Block struct {

	Index 		uint64		`json:"Index"`
	Timestamp 	int64		`json:"Timestamp"`
	Data 		string		`json:"Data"`
	PrevHash 	string		`json:"PrevHash"`
	BlockHash 	string		`json:"BlockHash"`

}

func (b* Block) GetHash() (bhash string) {

	record := strconv.FormatUint(b.Index, 16) +
			  strconv.FormatInt(b.Timestamp, 16) +
			  b.Data +
		      b.PrevHash

	h := crypto.SHA256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)

	return hex.EncodeToString(hashed)
}

func (b* Block) DecodeFromHex(hex string) (err error) {

	bs, err := hexutil.Decode(hex)

	if err != nil {
		return err
	}

	if err = json.Unmarshal(bs, b); err != nil {
		return err
	}

	return nil
}

func (b* Block) EncodeToHex() (hex string, err error) {

	bs, err := json.Marshal(b)

	if err != nil {
		return "", err
	}

	return hexutil.Encode(bs),nil
}

type Chain struct {
	blocks list.List
}

//建立新链
func (c *Chain) GenerateChain(datapath string) {

	genBlock := Block{1,time.Now().Unix(), datapath,"",""}
	genBlock.BlockHash = genBlock.GetHash()

	c.blocks.PushBack(&genBlock)

	return
}

func (c *Chain) BlockNumber() (uint64) {

	if c.blocks.Len() <= 0 {
		return 0
	} else {
		return c.LatestBlock().Index
	}
}

func (c *Chain) LatestBlock() (b *Block) {
	return c.blocks.Back().Value.(*Block)
}

func (c *Chain) GenerateNewBlock(datapath string){

	if c.BlockNumber() <= 0 {
		c.GenerateChain(datapath)
	} else {
		newBlock := Block{c.LatestBlock().Index + 1, time.Now().Unix(), datapath, c.LatestBlock().BlockHash, ""}
		newBlock.BlockHash = newBlock.GetHash()
		c.blocks.PushBack(&newBlock)
	}

}

func (c *Chain) AppendBlock(bl *Block) {

	c.blocks.PushBack(bl)
}

func (c *Chain) DumpPrint() {

	for it := c.blocks.Front(); it != nil; it = it.Next() {

		bytes, err := json.MarshalIndent((*it.Value.(*Block)), "", "  ")

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s\n", string(bytes) )
	}
}