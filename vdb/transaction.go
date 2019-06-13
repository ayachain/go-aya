package vdb

import (
	"context"
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
	lockers map[string]*sync.RWMutex

}

func (t *Transaction) Commit() error {

	ctx := WithCancel(context.TODO())

	var waiterLock sync.WaitGroup

	for k, vtx := range t.transactions {

		waiterLock.Add(1)

		go func( commitTx *leveldb.Transaction, lock *sync.RWMutex ) {

			lock.RLock()
			defer lock.RUnlock()

			err := commitTx.Commit()
			waiterLock.Done()

			if err != nil {
				ctx.CancelWithErr(err)
				return
			}

			select {
			case <- ctx.Done():

				if ctx.HasError() != nil {
					commitTx.Discard()
				}

				return
			}

		}( vtx, t.lockers[k] )

	}

	waiterLock.Wait()

	if ctx.HasError() != nil {
		return ErrCommitRolledBack
	}

	ctx.CancelWithErr(nil)

	fmt.Println("\nCommitSuccess")

	return nil
}

func (t *Transaction) Discard() {

	var waiterLock sync.WaitGroup

	for k, tx := range t.transactions {

		waiterLock.Add(1)
		go func( tx *leveldb.Transaction, mutex *sync.RWMutex ) {

			mutex.RLock()
			defer func() {
				mutex.Unlock()
				waiterLock.Done()
			}()

			tx.Discard()

		}(tx, t.lockers[k])

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