package common

import (
	"context"
	"time"
)

type Signal string

const (
	SignalInterrupt		Signal	= "Interrupt"
	SignalObserved 		Signal	= "Observed"
	SignalDoPacking		Signal	= "DoPacking"
	SignalDoMining		Signal  = "DoMining"
	SignalDoReceipting	Signal 	= "DoReceipting"
	SignalDoConfirming	Signal	= "DoConfirming"
	SignalOnConfirmed	Signal 	= "OnConfirmed"
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
