package receipt

import (
	"context"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-ipfs/core"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type aReceipt struct {

	Services

	ind *core.IpfsNode

	idxs indexes.IndexesServices
}

func CreateServices( ind *core.IpfsNode, idxServices indexes.IndexesServices ) Services {

	return &aReceipt{
		ind:ind,
		idxs:idxServices,
	}

}

func (r *aReceipt) HasTransactionReceipt( txhs EComm.Hash, idx ... *indexes.Index ) bool {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = r.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), r.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	exist := false
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		it := db.NewIterator( util.BytesPrefix(txhs.Bytes()), nil )
		defer it.Release()

		if it.Next() {
			exist = true
		}
		return nil

	}, DBPath); err != nil {
		log.Error(err)
	}

	return exist
}

func (r *aReceipt) GetTransactionReceipt( txhs EComm.Hash, idx ... *indexes.Index ) (*im.Receipt, error) {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = r.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}


	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), r.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	rp := &im.Receipt{}
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		it := db.NewIterator( util.BytesPrefix(txhs.Bytes()), nil )
		defer it.Release()

		if it.Next() {

			if err := proto.Unmarshal(it.Value(), rp); err != nil {
				return err
			}

		}

		return nil

	}, DBPath); err != nil {
		log.Error(err)
	}

	return rp, nil
}


func (r *aReceipt) NewWriter() (AVdbComm.VDBCacheServices, error) {
	return newWriter(r)
}