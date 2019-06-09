package common

import "errors"

var (
	ErrExpected 			= errors.New("lock or unlock amount expected")
	ErrTxVerifyExpected 	= errors.New("transaction verify failed")
	ErrParmasExpected		= errors.New("transaction params expected")
	ErrMessageTypeExped  	= errors.New("message type expected")

	TxConfirm              	= []byte("Confirm")
	TxReceiptsNotenough 	= []byte("not enough avail or voting balance")
)