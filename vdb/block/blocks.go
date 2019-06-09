package block

import (
	"errors"
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/headers"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)


var BestBlockKey = []byte("_BestBlock")

type aBlocks struct {
	BlocksAPI
	*mfs.Directory

	headAPI headers.HeadersAPI
	rawdb *leveldb.DB

	RWLocker sync.RWMutex
}

func CreateServices( mdir *mfs.Directory, hapi headers.HeadersAPI) BlocksAPI {

	api := &aBlocks{
		Directory:mdir,
		headAPI:hapi,
	}

	api.rawdb = common.OpenExistedDB(mdir, DBPath)

	return api
}

func (blks *aBlocks) GetBlocks( hashOrIndex...interface{} ) ([]*Block, error) {

	var blist []*Block

	for _, v := range hashOrIndex {

		var bhash EComm.Hash

		switch v.(type) {

		case uint64:

			hd, err := blks.headAPI.HeaderOf(v.(uint64))
			if err != nil {
				return nil, err
			}
			bhash = hd.Hash

		case EComm.Hash:
			bhash = v.(EComm.Hash)

		default:
			return nil, errors.New("input params must be a index(uint64) or cid object")
		}

		dbval, err := blks.rawdb.Get( bhash.Bytes(), nil )
		if err != nil {
			return nil, err
		}

		subBlock := &Block{}
		if err := subBlock.Decode(dbval); err != nil {
			return nil, err
		}

		blist = append(blist, subBlock)
	}

	return blist, nil
}

func (blks *aBlocks) BestBlock() *Block {

	kbs, err := blks.rawdb.Get(BestBlockKey, nil)
	if err != nil {
		return nil
	}

	blkbs, err := blks.rawdb.Get(kbs, nil)
	if err != nil {
		return nil
	}

	bestBlock := &Block{}
	if err := bestBlock.Decode(blkbs); err != nil {
		return nil
	}

	return bestBlock
}

func (blks *aBlocks) OpenVDBTransaction() (*leveldb.Transaction, *sync.RWMutex, error) {

	tx, err := blks.rawdb.OpenTransaction()
	if err != nil {
		return nil, nil, err
	}

	return tx, &blks.RWLocker, nil
}

func (blks *aBlocks) Close() {

	blks.RWLocker.Lock()
	defer blks.RWLocker.Unlock()

	_ = blks.rawdb.Close()

}

func (blks *aBlocks) AppendBlocks( group *AWork.TaskBatchGroup, blocks...*Block ) error {

	if len(blocks) <= 0 {
		return nil
	}

	var latesthash EComm.Hash

	for _, v := range blocks {

		latesthash = v.GetHash()

		rawvalue := v.Encode()

		group.Put( DBPath, latesthash.Bytes(), rawvalue )

	}

	group.Put(DBPath, []byte(BestBlockKey), latesthash.Bytes())

	return nil
}


func (blks *aBlocks) WriteGenBlock( group *AWork.TaskBatchGroup, gen *GenBlock ) error {


	hash := gen.GetHash().Bytes()

	group.Put( DBPath, hash, gen.Encode() )

	group.Put(DBPath, []byte(BestBlockKey), hash)

	return nil

}