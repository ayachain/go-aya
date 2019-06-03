package common

import (
	"bytes"
	"encoding/json"
	ABlock "github.com/ayachain/go-aya/block"
)

func BlockEqual( a *ABlock.Block, b *ABlock.Block ) bool {

	abs, _ := json.Marshal(a)
	bbs, _ := json.Marshal(b)

	return bytes.Equal(abs, bbs)
}