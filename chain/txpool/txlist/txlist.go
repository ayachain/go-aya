package txlist

import (
	"container/list"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	"github.com/pkg/errors"
	"sync"
)

var (
	ErrTIDIsAlReadyExist = errors.New("tx hash already exist in pool")
)

type TxList struct {

	list *list.List

	wmu sync.Mutex

}

func NewTxList( tx *ATx.Transaction ) *TxList {

	l := list.New()

	l.PushBack(tx)

	return &TxList{
		list:l,
	}

}

func (l *TxList) Len() int {

	l.wmu.Lock()
	defer l.wmu.Unlock()

	return l.list.Len()

}

func (l *TxList) FrontTx() *ATx.Transaction {

	l.wmu.Lock()
	defer l.wmu.Unlock()

	if l.list.Len() == 0 {
		return nil
	}

	return l.list.Front().Value.(*ATx.Transaction)
}


func (l *TxList) PopFront() {

	l.wmu.Lock()
	defer l.wmu.Unlock()

	if l.list.Len() > 0 {
		l.list.Remove(l.list.Front())
	}

}

func (l *TxList) RemoveFromTid( tid uint64 ) bool {

	for i := l.list.Front(); i != nil; i = i.Next() {

		if i.Value.(*ATx.Transaction).Tid == tid {

			l.list.Remove(i)
			return true

		}

	}

	return false

}

func (l *TxList) GetLinearTxsFromFront() []*ATx.Transaction {

	l.wmu.Lock()
	defer l.wmu.Unlock()

	if l.list.Len() <= 0 {
		return nil
	}

	var txs []*ATx.Transaction

	stid := l.list.Front().Value.(*ATx.Transaction).Tid

	for i := l.list.Front(); i != nil; i = i.Next() {

		if i.Value.(*ATx.Transaction).Tid == stid {
			txs = append(txs, i.Value.(*ATx.Transaction))
			stid ++

		} else {

			break

		}

	}

	return txs
}

func (l *TxList) AddTx( transaction *ATx.Transaction ) error {

	l.wmu.Lock()
	defer l.wmu.Unlock()

	if l.list.Len() == 0 {

		l.list.PushBack( transaction )

		return nil

	} else {

		for i := l.list.Front(); i != nil; i = i.Next() {

			if i.Value.(*ATx.Transaction).Tid == transaction.Tid {

				return ErrTIDIsAlReadyExist

			} else if i.Value.(*ATx.Transaction).Tid > transaction.Tid {

				l.list.InsertAfter( transaction, i )

				return nil

			} else {

				if i == l.list.Back() {

					l.list.PushBack(transaction)

					return nil

				} else {

					continue

				}

			}

		}

	}

	return nil
}