package headers

import (
	EComm "github.com/ethereum/go-ethereum/common"
)

func HeadersToHash( hs []*Header ) []EComm.Hash {

	var hashs []EComm.Hash

	for _, v := range hs {

		hashs = append(hashs, v.Hash)

	}

	return hashs
}