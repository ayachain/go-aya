package vdb

import (
	"fmt"
	"github.com/ayachain/go-aya/consensus/core/worker"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"sync"
)

var (
	ErrDBTargetNotExist = errors.New("transaction target db not exist")
	ErrCommitRolledBack = errors.New("submission rolled back" )
)

type Transaction struct {
	transactions map[string]*leveldb.Transaction
}

func (t *Transaction) Commit() error {

	var waiterLock sync.WaitGroup

	errs := &sync.Map{}

	for k, vtx := range t.transactions {

		waiterLock.Add(1)

		go func( commitTx *leveldb.Transaction ) {

			errs.Store(k, commitTx.Commit())

			waiterLock.Done()

		}( vtx )

	}

	waiterLock.Wait()

	needRollBack := false
	errs.Range(func(key, value interface{}) bool {

		if value != nil {
			needRollBack = true
			return false
		}

		return true
	})

	if needRollBack {

		var rollbackgroup sync.WaitGroup

		for _, vtx := range t.transactions {

			rollbackgroup.Add(1)

			go func(commitTx *leveldb.Transaction) {

				commitTx.Discard()

				rollbackgroup.Done()

			}(vtx)

		}

		log.Error( ErrCommitRolledBack )

		return ErrCommitRolledBack
	}


	return nil
}

func (t *Transaction) Discard() {

	var waiterLock sync.WaitGroup

	for _, tx := range t.transactions {

		waiterLock.Add(1)
		go func( tx *leveldb.Transaction ) {

			tx.Discard()

		}(tx)

	}

	waiterLock.Wait()

	fmt.Println("DiscardSuccess")
}

func (t *Transaction) Write( group *worker.TaskBatchGroup ) error {

	bmap := group.GetBatchMap()

	for k, batch := range bmap {

		tx, exist := t.transactions[k]
		if !exist {
			return ErrDBTargetNotExist
		}

		if err := tx.Write(batch, &opt.WriteOptions{Sync:true}); err != nil {
			return nil
		}

	}

	return nil
}