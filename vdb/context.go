package vdb

import "context"

type TransactionContext struct {
	context.Context
	cancel context.CancelFunc
	e error
	canceled bool
}


func WithCancel( ctx context.Context ) *TransactionContext {

	ctx, cancel := context.WithCancel(ctx)

	return &TransactionContext{
		Context:ctx,
		cancel:cancel,
		e:nil,
		canceled:false,
	}

}

func (ttxc *TransactionContext) CancelWithErr( err error ) {

	if ttxc.canceled {
		return
	}

	ttxc.canceled = true
	ttxc.e = err
	ttxc.cancel()
}


func (ttxc *TransactionContext) HasError() error {
	return ttxc.e
}