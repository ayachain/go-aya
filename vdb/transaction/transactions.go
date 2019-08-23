package transaction

import (
	"context"
	"encoding/binary"
	"fmt"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/im"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-ipfs/core"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (

	TxOutCountPrefix = []byte("TOC_")

	TxTotalCountPrefix = []byte("TTC_")

	TxHistoryPrefix = []byte("TH_")

)

type aTransactions struct {

	Services

	ind *core.IpfsNode

	idxs indexes.IndexesServices
}

func CreateServices( ind *core.IpfsNode, idxServices indexes.IndexesServices ) Services {

	return &aTransactions{
		ind:ind,
		idxs:idxServices,
	}

}
func (txs *aTransactions) NewWriter() (AVdbComm.VDBCacheServices, error) {
	return newWriter(txs)
}

func (txs *aTransactions) GetTxCount( address EComm.Address, idx ... *indexes.Index ) (uint64, error) {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = txs.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}


	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), txs.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	total := uint64(0)
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		key := append(TxOutCountPrefix, address.Bytes()... )

		v, err := db.Get(key, nil)
		if err != nil {

			if err == leveldb.ErrNotFound {
				return nil
			} else {
				return err
			}

		}

		total = binary.BigEndian.Uint64(v)
		return nil

	}, DBPath); err != nil {

		log.Error(err)
		return total, err
	}

	return total, nil
}


func (txs *aTransactions) GetTxByHash( hash EComm.Hash, idx ... *indexes.Index ) (*im.Transaction, error) {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = txs.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), txs.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	tx := &im.Transaction{}
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		it := db.NewIterator( util.BytesPrefix(hash.Bytes()) , nil)
		defer it.Release()

		if it.Next() {

			if err := proto.Unmarshal(it.Value(), tx); err != nil {
				return fmt.Errorf("%v can't found transaction", hash.String())
			}

		} else {

			return fmt.Errorf("%v can't found transaction", hash.String())

		}

		return nil

	}, DBPath); err != nil {
		return nil, err
	}

	return tx, nil
}


func (txs *aTransactions) GetTxByHashBs( hsbs []byte, idx ... *indexes.Index ) (*im.Transaction, error) {

	hash := EComm.BytesToHash(hsbs)

	return txs.GetTxByHash(hash, idx...)
}

func (txs *aTransactions) GetHistoryHash( address EComm.Address, offset uint64, size uint64, idx ... *indexes.Index) []EComm.Hash {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = txs.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), txs.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	itkey := append(TxHistoryPrefix, address.Bytes()...)

	var hashs []EComm.Hash
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		it := db.NewIterator( util.BytesPrefix(itkey), nil )
		defer it.Release()

		if offset > 0 {

			seekKey := append(itkey, AVdbComm.BigEndianBytes(offset - 1)...)
			if !it.Seek(seekKey) {
				return errors.New("seek history key expected ")
			}
		}

		s := uint64(0)
		for it.Next() {

			if s > size - 1 {
				break
			}

			hashs = append(hashs, EComm.BytesToHash(it.Value()[:32]))

			s ++
		}

		return nil

	}, DBPath); err != nil {
		return nil
	}

	return hashs
}

func (txs *aTransactions) GetHistoryContent( address EComm.Address, offset uint64, size uint64, idx ... *indexes.Index) ([]*im.Transaction, error) {

	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = txs.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), txs.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	var tlist []*im.Transaction
	if err := ADB.ReadClose( dbroot, func(db *leveldb.DB) error {

		itkey := append(TxHistoryPrefix, address.Bytes()...)

		it := db.NewIterator( util.BytesPrefix(itkey), nil )
		defer it.Release()

		if offset > 0 {

			seekKey := append(itkey, AVdbComm.BigEndianBytes(offset - 1)...)

			if !it.Seek(seekKey) {
				return errors.New("seek history key expected ")
			}
		}

		s := uint64(0)
		for it.Next() {

			if s > size - 1 {
				break
			}

			bs, err := db.Get(it.Value(), nil)
			if err != nil {
				return err
			}

			tx := &im.Transaction{}

			if err := proto.Unmarshal(bs, tx); err != nil {
				return err
			}

			tlist = append(tlist, tx)
			s ++
		}

		return nil

	}, DBPath); err != nil {
		return nil, err
	}

	return tlist, nil
}