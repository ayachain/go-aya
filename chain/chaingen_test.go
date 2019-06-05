package chain

import (
	"fmt"
	"github.com/ayachain/go-aya/vdb/block"
	"testing"
)

var ChainGenJson = `
{
	"index": 0,
	"chainid": "aya",
	"parent": "f919678e3089f1da5b4eac138c86efa0461555223f71f0d53a5a1ea9472834a0",
	"timestamp": 1501516800,
	"append": "SSUyMHJlc2VydmVkJTIwYW4lMjBpbXBvcnRhbnQlMjBtZXNzYWdlJTIwaW4lMjBwYXJlbnQlMjBoYXNoLg==",
	"extradata": "",
	"txc": 0,
	"txs": "",
	"award": {
		"0xD2bfC9AC49F3F0CfC1cBDF1cf579593D6De85435": 1000000000
	}
}
`


func TestGenChaiDB(t *testing.T) {


	genBlock := &block.GenBlock{}

	if err := genBlock.Decode([]byte(ChainGenJson)); err != nil {
		t.Fatal("gen block unmarshal expected")
	}

	fmt.Printf("%v", genBlock.AppendData)
}