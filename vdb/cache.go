package vdb

import (
	AAssetses "github.com/ayachain/go-aya/vdb/assets"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ayachain/go-aya/vdb/common"
	AVDBComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ayachain/go-aya/vdb/merger"
	ANodes "github.com/ayachain/go-aya/vdb/node"
	AReceipts "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/ipfs/go-cid"
)

type CacheCVFS interface {

	Close() error

	BestCID() cid.Cid

	Blocks() ABlock.MergeWriter

	Nodes() ANodes.MergeWriter

	Assetses() AAssetses.MergeWriter

	Receipts() AReceipts.MergeWriter

	Transactions() ATx.MergeWriter

	MergeGroup() merger.CVFSMerger
}


type aCacheCVFS struct {

	CacheCVFS

	rdonlyCVFS	CVFS

	cacheSers map[string]interface{}
}

func NewCacheCVFS( rdonlyCVFS CVFS ) (*aCacheCVFS, error) {

	//var err error
	cache := &aCacheCVFS{
		rdonlyCVFS : rdonlyCVFS,
		cacheSers:make(map[string]interface{}),
	}

	return cache, nil
}

func (cache *aCacheCVFS) Close() error {

	for _, db := range cache.cacheSers {

		vdbs, ok := db.(common.VDBCacheServices)
		if ok {
			vdbs.Close()
		}

	}

	return nil

}

func (cache *aCacheCVFS) Nodes() ANodes.MergeWriter {

	var err error
	ser, exist := cache.cacheSers[ANodes.DBPath]
	if !exist {

		ser, err = cache.rdonlyCVFS.Nodes().NewWriter()
		if err != nil {
			return nil
		}

		cache.cacheSers[ANodes.DBPath] = ser
	}


	wt, ok := ser.(ANodes.MergeWriter)

	if !ok {
		return nil
	}

	return wt

}

func (cache *aCacheCVFS) Blocks() ABlock.MergeWriter {

	var err error
	ser, exist := cache.cacheSers[ABlock.DBPath]
	if !exist {

		ser, err = cache.rdonlyCVFS.Blocks().NewWriter()
		if err != nil {
			return nil
		}

		cache.cacheSers[ABlock.DBPath] = ser
	}


	wt, ok := ser.(ABlock.MergeWriter)

	if !ok {
		return nil
	}

	return wt
}

func (cache *aCacheCVFS) Assetses() AAssetses.MergeWriter {

	var err error
	ser, exist := cache.cacheSers[AAssetses.DBPath]
	if !exist {

		ser, err = cache.rdonlyCVFS.Assetses().NewWriter()
		if err != nil {
			return nil
		}

		cache.cacheSers[AAssetses.DBPath] = ser
	}

	wt, ok := ser.(AAssetses.MergeWriter)

	if !ok {
		return nil
	}

	return wt
}

func (cache *aCacheCVFS) Receipts() AReceipts.MergeWriter {

	var err error
	ser, exist := cache.cacheSers[AReceipts.DBPath]
	if !exist {

		ser, err = cache.rdonlyCVFS.Receipts().NewWriter()
		if err != nil {
			return nil
		}

		cache.cacheSers[AReceipts.DBPath] = ser
	}

	wt, ok := ser.(AReceipts.MergeWriter)

	if !ok {
		return nil
	}

	return wt
}

func (cache *aCacheCVFS) Transactions() ATx.MergeWriter {

	var err error
	ser, exist := cache.cacheSers[ATx.DBPath]
	if !exist {

		ser, err = cache.rdonlyCVFS.Transactions().NewWriter()
		if err != nil {
			return nil
		}

		cache.cacheSers[ATx.DBPath] = ser

	}

	wt, ok := ser.(ATx.MergeWriter)

	if !ok {
		return nil
	}

	return wt
}

func (cache *aCacheCVFS) MergeGroup() merger.CVFSMerger {

	merger := merger.NewMerger()

	for _, path := range AVDBComm.StorageDBPaths {

		if db, exist := cache.cacheSers[path]; exist {

			vdbs, ok := db.(common.VDBCacheServices)

			if ok {

				batch := vdbs.MergerBatch()

				if batch.Len() > 0 {

					merger.GetBatchMap()[path] = batch

				} else {

					merger.GetBatchMap()[path] = nil
				}

			}

		}

	}

	return merger
}