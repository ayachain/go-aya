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
}

func (t *Transaction) Commit() error {

	ctx := WithCancel(context.TODO())

	var waiterLock sync.WaitGroup

	for _, vtx := range t.transactions {

		waiterLock.Add(1)

		go func( commitTx *leveldb.Transaction ) {

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

		}( vtx )

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