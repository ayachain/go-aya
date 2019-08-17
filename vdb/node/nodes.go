package node

import (
	"context"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	"github.com/ipfs/go-ipfs/core"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type aNodes struct {

	Services

	ind *core.IpfsNode

	idxs indexes.IndexesServices
}

func CreateServices( ind *core.IpfsNode, idxServices indexes.IndexesServices ) Services {

	return &aNodes{
		ind:ind,
		idxs:idxServices,
	}

}

func (api *aNodes) NewWriter() (AVdbComm.VDBCacheServices, error) {
	return newWriter( api )
}

func (api *aNodes) GetNodeByPeerId( peerId string, idx ... *indexes.Index ) (*Node, error) {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = api.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), api.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	nd := &Node{}
	if err := ADB.ReadClose(dbroot, func(db *leveldb.DB) error {

		bs, err := db.Get( []byte(peerId), nil )
		if err != nil {
			return err
		}

		if err := nd.Decode(bs); err != nil {
			return err
		}
		return nil

	}, DBPath); err != nil {

		return nil, err
	}

	return nd, nil
}

func (api *aNodes) GetSuperNodeList( idx ... *indexes.Index ) []*Node {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = api.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), api.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	var rets[] *Node
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		it := db.NewIterator( util.BytesPrefix( []byte(NodeTypeSuper) ), nil )

		defer it.Release()

		for it.Next() {

			perrId := it.Value()
			
			bs, err := db.Get(perrId, nil)
			if err != nil {
				return err
			}

			nd := &Node{}
			if err := nd.Decode(bs); err == nil {
				rets = append(rets, nd)
			}
		}

		return nil

	}, DBPath); err != nil {
		return nil
	}

	return rets
}

func (api *aNodes) GetSuperMaterTotalVotes( idx ... *indexes.Index ) uint64 {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = api.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), api.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	var total uint64
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		it := db.NewIterator( util.BytesPrefix( []byte(NodeTypeSuper) ), nil )
		defer it.Release()

		for it.Next() {

			perrId := it.Value()

			bs, err := db.Get(perrId, nil)
			if err != nil {
				panic(err)
			}

			nd := &Node{}
			if err := nd.Decode(bs); err == nil {
				total += nd.Votes
			}
		}

		return nil

	}, DBPath); err != nil {
		log.Error(err)
	}

	return total
}

func (api *aNodes) GetFirst( idx ... *indexes.Index ) *Node {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = api.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), api.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	nd := &Node{}
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		if bs, err := db.Get( []byte("Super00000001"), nil ); err != nil {

			return err

		} else {

			if err := nd.Decode(bs); err != nil {
				return err
			}

			return nil
		}

	}, DBPath); err != nil {
		return nil
	}

	return nd
}

func (api *aNodes) GetSuperNodeCount( idx ... *indexes.Index ) int64 {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = api.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), api.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	var s = int64(0)
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		it := db.NewIterator( util.BytesPrefix([]byte(NodeTypeSuper)), nil)
		defer it.Release()

		for it.Next() {
			s ++
		}

		return nil

	},DBPath); err != nil {
		log.Error(err)
	}

	return s
}

func (api *aNodes) DoRead( readingFunc ADB.ReadingFunc, idx ... *indexes.Index ) error {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = api.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), api.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	return ADB.ReadClose( dbroot, readingFunc, DBPath )
}