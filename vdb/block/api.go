package block

import (
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
)

const DBPath = "/db/blocks"

type BlocksAPI interface {
	AVdbComm.VDBSerices
	GetBlocks ( iorc... interface{} ) ([]*Block, error)
}