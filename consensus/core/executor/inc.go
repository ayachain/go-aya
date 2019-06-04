package executor

import (
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	"github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ethereum/go-ethereum/common"
)

//
// Finally, the object that executes the transaction can refer to writing "log" log locally,
// and "TaskGroup" object can be saved as a file by binary serialization, and then executed
// at the beginning, which guarantees the atomicity of the writing, and can also be restored
// in unexpected circumstances.
//
type Executor interface {
	ACStep.ConsensusOver

	Exec( group *worker.TaskBatchGroup ) (error, func( hash common.Hash ))

	ResolverLogs( ) error
}