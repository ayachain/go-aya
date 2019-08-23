package minerpool

import (
	VDB "github.com/ayachain/go-aya/vdb"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/merger"
	"github.com/pkg/errors"
)

var (
	ErrNotLinearBlock	= errors.New("not a linear mining block")
	ErrInvalidLatest	= errors.New("invalid latest block")
	ErrInvalidChainID	= errors.New("invalid ChainID")
	ErrCreateBatch   	= errors.New("create cvfs merger failed")
	ErrReadTxsTimeOut	= errors.New("read transaction list from IPFS dag timeout")
	ErrLinkCVFS			= errors.New("can not link to target CVFS")
	ErrCreateCVFSCache  = errors.New("can not create cvfs cache(writer)")
)

type MiningTask struct {

	MiningBlock *im.Block

	Txs			[]*im.Transaction

	VWriter 	VDB.CacheCVFS
}

func NewTask( block *im.Block ) *MiningTask {
	return &MiningTask{
		MiningBlock:block,
	}
}

type MiningResult struct {

	Err error

	Batcher merger.CVFSMerger

	Task *MiningTask
}
