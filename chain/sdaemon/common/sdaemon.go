package common

import (
	"context"
	"time"
)

type Signal int8

const (
	SignalInterrupt		Signal	= -1
	SignalObserved 		Signal	= 0
	SignalDoPacking		Signal	= 1
	SignalDoMining		Signal  = 2
	SignalDoReceipting	Signal 	= 3
	SignalDoConfirming	Signal	= 4
	SignalOnConfirmed	Signal 	= 5
)

type TimeoutConfig struct {

	PackingDuration	time.Duration

	MiningDuration time.Duration

	ReceiptDuration time.Duration

	ConfirmDuration time.Duration
}

type StatDaemon interface {

	AddTimeoutObserver( timeoutFunc func ( s Signal ) ) error

	RemoveObserver( timeoutFunc func ( s Signal ) )

	PowerOn( context.Context )

	SendingSignal( bindex uint64, s Signal )
}
