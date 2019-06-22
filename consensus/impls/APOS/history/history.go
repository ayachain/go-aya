package history

import "github.com/ayachain/go-aya/vdb/node"

type History struct {

	mdb map[string] map[string]uint64

}

func New() *History {

	return &History{
		mdb: map[string]map[string]uint64{},
	}

}

func (h *History) CanConsensus ( hash string, node *node.Node, threshold uint64 ) bool {

	submap, exist := h.mdb[hash]

	if exist {

		_, pexist := submap[node.PeerID]

		if pexist {
			return false
		}

		submap[node.PeerID] += node.Votes

		if submap[node.PeerID] > threshold {

			delete(h.mdb, hash)

			return true
		}

		return false

	} else {

		h.mdb[hash][node.PeerID] = node.Votes

		return false

	}

}