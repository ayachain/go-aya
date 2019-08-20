package sdaemon

import (
	"context"
	"errors"
	"github.com/ayachain/go-aya/chain/sdaemon/common"
	"github.com/prometheus/common/log"
	"sync"
	"time"
)

var (
	ErrObserverAlreadyExists 	= errors.New("observer function already exist")

	DefaultConfig				= &common.TimeoutConfig {
		PackingDuration:	time.Second * 8,
		MiningDuration:		time.Second * 12,
		ReceiptDuration:	time.Second * 8,
		ConfirmDuration:	time.Second * 8,
	}
)

type aStatDaemon struct {

	common.StatDaemon

	tcnf *common.TimeoutConfig

	waitingTimer *time.Timer

	observers []func( s common.Signal )

	lsig common.Signal

	smu sync.Mutex

	attachBlockIndex uint64
}

func NewDaemon( cnf *common.TimeoutConfig ) common.StatDaemon {

	d := &aStatDaemon{
		observers:make([]func( s common.Signal ), 0),
		waitingTimer:time.NewTimer(0),
		tcnf:cnf,
	}

	d.waitingTimer.Stop()

	return d
}

func (d *aStatDaemon) RemoveObserver( timeoutFunc func( s common.Signal ) ) {

	for i, f := range d.observers {

		if &f == &timeoutFunc {
			d.observers[i] = d.observers[len(d.observers) - 1]
			d.observers = d.observers[:len(d.observers) - 1]
		}

	}

	return
}

func (d *aStatDaemon) AddTimeoutObserver( timeoutFunc func( s common.Signal ) ) error {

	for _, f := range d.observers {

		if &f == &timeoutFunc {
			return ErrObserverAlreadyExists
		}
	}

	d.observers = append( d.observers, timeoutFunc )

	timeoutFunc(common.SignalObserved)

	return nil
}

func (d *aStatDaemon) PowerOn(ctx context.Context) {

	log.Info("ASD On")
	defer func() {
		d.observers = make([]func( s common.Signal ), 0)
		log.Info("ASD Off")
	}()

	for {

		select {
		case <- ctx.Done():
			return

		case <- d.waitingTimer.C:
			log.Infof("%v Waiting timeout", d.lsig)
			break
		}

		for _, f := range d.observers {
			go f( d.lsig )
		}
	}
}

func (d *aStatDaemon) SendingSignal( bindex uint64, signal common.Signal ) {

	d.smu.Lock()
	defer d.smu.Unlock()

	d.lsig = signal
	if signal == common.SignalObserved {
		return
	}

	if signal == common.SignalDoPacking {

		d.attachBlockIndex = bindex
		d.waitingTimer.Reset(d.tcnf.PackingDuration)

		return

	} else if d.attachBlockIndex != bindex {

		return
	}

	switch signal {

	case common.SignalDoMining:
		d.waitingTimer.Reset(d.tcnf.MiningDuration)

	case common.SignalDoReceipting:
		d.waitingTimer.Reset(d.tcnf.ReceiptDuration)

	case common.SignalDoConfirming:
		d.waitingTimer.Reset(d.tcnf.ConfirmDuration)

	case common.SignalOnConfirmed:

		d.waitingTimer.Stop()

		for _, f := range d.observers {
			go f( common.SignalOnConfirmed )
		}

	case common.SignalInterrupt:
		d.waitingTimer.Stop()

	}

	return
}