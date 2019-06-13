package block

import (
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
)

const DBPath = "/db/blocks"

type BlocksAPI interface {

	AVdbComm.VDBSerices

	GetBlocks( hashOrIndex...interface{} ) ([]*Block, error)

	BestBlock()	*Block

	AppendBlocks( group *AWork.TaskBatchGroup, blocks...*Block ) error

	WriteGenBlock( group *AWork.TaskBatchGroup, gen *GenBlock ) error

}
