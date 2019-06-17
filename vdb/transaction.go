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

	errs := make(map[string]error)

	for k, vtx := range t.transactions {

		waiterLock.Add(1)

		go func( commitTx *leveldb.Transaction ) {

			errs[k] = commitTx.Commit()

			waiterLock.Done()

		}( vtx )

	}

	waiterLock.Wait()


	for _, err := range errs {

		if err != nil {
			goto rollback
		}

	}

	fmt.Println("\nCommitSuccess")

	return nil

rollback :

	var rollbackgroup sync.WaitGroup

	for _, vtx := range t.transactions {

		rollbackgroup.Add(1)

		go func( commitTx *leveldb.Transaction ) {

			commitTx.Discard()

			rollbackgroup.Done()

		}( vtx )

	}

	return ErrCommitRolledBack
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