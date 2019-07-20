package vdb

import (
	AWroker "github.com/ayachain/go-aya/consensus/core/worker"
	AAssetses "github.com/ayachain/go-aya/vdb/assets"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	"github.com/ayachain/go-aya/vdb/common"
	AReceipts "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
)

type CacheCVFS interface {

	Close() error

	Blocks() ABlock.Caches
	Assetses() AAssetses.Caches
	Receipts() AReceipts.Caches
	Transactions() ATx.Caches
	MergeGroup() *AWroker.TaskBatchGroup
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

func (cache *aCacheCVFS) Blocks() ABlock.Caches {

	var err error
	ser, exist := cache.cacheSers[ABlock.DBPath]
	if !exist {

		ser, err = cache.rdonlyCVFS.Blocks().NewCache()
		if err != nil {
			return nil
		}

		cache.cacheSers[ABlock.DBPath] = ser
	}


	wt, ok := ser.(ABlock.Caches)

	if !ok {
		return nil
	}

	return wt
}

func (cache *aCacheCVFS) Assetses() AAssetses.Caches {

	var err error
	ser, exist := cache.cacheSers[AAssetses.DBPath]
	if !exist {

		ser, err = cache.rdonlyCVFS.Assetses().NewCache()
		if err != nil {
			return nil
		}

		cache.cacheSers[AAssetses.DBPath] = ser
	}

	wt, ok := ser.(AAssetses.Caches)

	if !ok {
		return nil
	}

	return wt
}

func (cache *aCacheCVFS) Receipts() AReceipts.Caches {

	var err error
	ser, exist := cache.cacheSers[AReceipts.DBPath]
	if !exist {

		ser, err = cache.rdonlyCVFS.Receipts().NewCache()
		if err != nil {
			return nil
		}

		cache.cacheSers[AReceipts.DBPath] = ser
	}

	wt, ok := ser.(AReceipts.Caches)

	if !ok {
		return nil
	}

	return wt
}


func (cache *aCacheCVFS) Transactions() ATx.Caches {

	var err error
	ser, exist := cache.cacheSers[ATx.DBPath]
	if !exist {

		ser, err = cache.rdonlyCVFS.Transactions().NewCache()
		if err != nil {
			return nil
		}

		cache.cacheSers[ATx.DBPath] = ser

	}

	wt, ok := ser.(ATx.Caches)

	if !ok {
		return nil
	}

	return wt
}

func (cache *aCacheCVFS) MergeGroup() *AWroker.TaskBatchGroup {

	group := AWroker.NewGroup()

	for k, db := range cache.cacheSers {

		vdbs, ok := db.(common.VDBCacheServices)
		if ok {
			group.GetBatchMap()[k] = vdbs.MergerBatch()
		}

	}

	return group
}