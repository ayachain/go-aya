package node

import (
	"github.com/ayachain/go-aya/vdb/im"
	"time"
)

type PingRet struct {
	Node *im.Node
	UTime int64
	RTT time.Duration
}
