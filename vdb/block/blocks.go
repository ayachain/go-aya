package block

import (
	"context"
	"encoding/binary"
	"errors"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/indexes"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-ipfs/core"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"strings"
)

var(
	LatestPosBlockIdxKey 	= []byte("LatestPosBlockIndex")
	ErrBlockNotFound		= errors.New("can not found block")
)


type aBlocks struct {

	reader

	ind *core.IpfsNode

	idxs indexes.IndexesServices
}

func CreateServices( ind *core.IpfsNode, idxServices indexes.IndexesServices ) Services {

	return &aBlocks{
		ind:ind,
		idxs:idxServices,
	}

}

func (blks *aBlocks) GetLatestPosBlockIndex( idx ... *indexes.Index ) uint64 {


	var lidx *indexes.Index
	var err error
	if len(idx) > 0 {

		lidx = idx[0]

	} else {

		lidx, err = blks.idxs.GetLatest()
		if err != nil {
			panic(err)
		}
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), blks.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	r := uint64(0)
	if err := ADB.ReadClose(dbroot, func(db *leveldb.DB) error {

		if exist, err := db.Has(LatestPosBlockIdxKey, nil); err != nil {
			return err

		} else if !exist {

			return leveldb.ErrNotFound

		} else {

			if bs, err := db.Get(LatestPosBlockIdxKey, nil); err != nil {
				return err

			} else {
				r = binary.BigEndian.Uint64(bs)
				return nil
			}
		}

	}, DBPath); err != nil {
		log.Error(err)
	}

	return r
}

func (blks *aBlocks) GetLatestBlock() (*Block, error) {

	blocks, err := blks.GetBlocks(BlockNameLatest)
	if err != nil {
		return nil, err
	}

	if blocks == nil || len(blocks) == 0 {
		return nil, ErrBlockNotFound
	}

	return blocks[0], nil
}

func (blks *aBlocks) GetBlocks( hashOrIndex...interface{} ) ([]*Block, error) {

	lidx, err := blks.idxs.GetLatest()
	if err != nil {
		panic(err)
	}

	dbroot, err, cls := AVdbComm.GetDBRoot(context.TODO(), blks.ind, lidx.FullCID, DBPath)
	if err != nil {
		panic(err)
	}
	defer cls()

	var hashList []EComm.Hash
	for _, v := range hashOrIndex {

		var bhash EComm.Hash

		switch v.(type) {

		case string:

			switch strings.ToLower(v.(string)) {
			case BlockNameLatest:

				hd, err := blks.idxs.GetLatest()
				if err != nil {
					return nil, err
				}
				bhash = hd.Hash


			case BlockNameGen:

				hd, err := blks.idxs.GetIndex(0)
				if err != nil {
					return nil, err
				}
				bhash = hd.Hash
			}

		case uint64:

			switch v.(uint64) {

			case ^uint64(0): /// Latest

				hd, err := blks.idxs.GetLatest()
				if err != nil {
					return nil, ErrBlockNotFound
				}
				bhash = hd.Hash

			default:

				hd, err := blks.idxs.GetIndex(v.(uint64))
				if err != nil {
					return nil, ErrBlockNotFound
				}

				bhash = hd.Hash
			}

		case EComm.Hash:
			bhash = v.(EComm.Hash)

		default:
			return nil, errors.New("input params must be a index(uint64) or cid object")
		}

		hashList = append(hashList, bhash)
	}

	var blist []*Block
	if err := ADB.ReadClose(dbroot, func(db *leveldb.DB) error {

		for _, bhash := range hashList {

			dbval, err := db.Get( bhash.Bytes(), nil )
			if err != nil {
				return ErrBlockNotFound
			}

			subBlock := &Block{}
			if err := subBlock.Decode(dbval); err != nil {
				return ErrBlockNotFound
			}

			blist = append(blist, subBlock)
		}

		return nil

	}, DBPath); err != nil {
		return nil, err
	}

	return blist, nil
}

func (blks *aBlocks) NewWriter() (AVdbComm.VDBCacheServices, error) {
	return newWriter(blks)
}