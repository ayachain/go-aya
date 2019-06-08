package block

import (
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	Avdb "github.com/ayachain/go-aya/vdb"
	"github.com/ipfs/go-ipfs/core"
)

func SignaturerStep ( msg interface{}, vdb Avdb.CVFS, ind *core.IpfsNode, group *AWork.TaskBatchGroup ) (interface{}, error) {
	return msg ,nil
}