package minerpool

import (
	VDB "github.com/ayachain/go-aya/vdb"
	AMBlock "github.com/ayachain/go-aya/vdb/mblock"
	"github.com/ayachain/go-aya/vdb/merger"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/pkg/errors"
)

var (
	ErrNotLinearBlock	= errors.New("not a linear mining block")
	ErrInvalidLatest	= errors.New("invalid latest block")
	ErrInvalidChainID	= errors.New("invalid ChainID")
	ErrContextCancel 	= errors.New("parent context cancel")
	ErrCreateBatch   	= errors.New("create cvfs merger failed")
	ErrReadTxsTimeOut	= errors.New("read transaction list from IPFS dag timeout")
	ErrReadIdxServices  = errors.New("can not use cid create index services")
	ErrLinkCVFS			= errors.New("can not link to target CVFS")
	ErrCreateCVFSCache  = errors.New("can not create cvfs cache(writer)")
	ErrWorkingTimeout	= errors.New("mining timeout")
)

type MiningTask struct {

	MiningBlock *AMBlock.MBlock

	Txs			[]*ATx.Transaction

	VWriter 	VDB.CacheCVFS
}

func NewTask( block *AMBlock.MBlock ) *MiningTask {
	return &MiningTask{
		MiningBlock:block,
	}
}

type MiningResult struct {

	Err error

	Batcher merger.CVFSMerger

	Task *MiningTask
}
