package block

import (
	"context"
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	Avdb "github.com/ayachain/go-aya/vdb"
	"github.com/ipfs/go-ipfs/core"
)

func ExecutorTransaction( msg interface{}, vdb Avdb.CVFS, ind *core.IpfsNode, group *AWork.TaskBatchGroup ) (interface{}, error) {

	// begin commit to main vdb, use transaction
	exeTx, err := vdb.OpenTransaction()
	if err != nil {
		return nil, err
	}

	// write batch to transaction
	if err := exeTx.Write(group); err != nil {
		return nil, err
	}

	// try commit, if commit has any error, it will roll back automatically
	if err := exeTx.Commit(); err != nil {
		return nil, err
	}

	// transaction success, try get latest vdb root path cid
	mcid, err := vdb.Flush(context.TODO())
	if err != nil {
		return nil, err
	}

	return mcid, err
}