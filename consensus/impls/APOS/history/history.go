package history

import (
	"github.com/ayachain/go-aya/vdb/node"
	"sync"
)

type History struct {

	mdb map[string] map[string]uint64

	mapmutex sync.Mutex

}

func New() *History {

	return &History{
		mdb: map[string]map[string]uint64{},
	}

}

func (h *History) Clear() {

	h.mapmutex.Lock()
	defer h.mapmutex.Unlock()

	h.mdb = map[string]map[string]uint64{}
}


func (h *History) CanConsensus ( hash string, node *node.Node, threshold uint64 ) bool {

	h.mapmutex.Lock()
	defer h.mapmutex.Unlock()

	submap, exist := h.mdb[hash]

	if exist {

		_, pexist := submap[node.PeerID]

		if pexist {
			return false
		}

		submap[node.PeerID] = node.Votes

		var totalVotes uint64 = 0

		for _, v := range submap {
			totalVotes += v
		}

		if totalVotes > threshold && len(submap) >= 3 {

			delete(h.mdb, hash)

			return true
		}

		return false

	} else {

		h.mdb[hash] = map[string]uint64{}

		h.mdb[hash][node.PeerID] = node.Votes

		return false

	}

}