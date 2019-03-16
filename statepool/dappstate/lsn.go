package dappstate

import (
	"github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
	"log"
)

const (
	ListennerThread_Stop 	= -1
	ListennerThread_Running = 0
	ListennerThread_Dead	= 2
)

type Listener interface {

	StartListening() error

	Shutdown() error

	ThreadState() int

	GetTopics() string
}

type BaseListner struct {

	Listener

	handleDelegate func (msg *shell.Message)

	state*	DappState

	threadstate int

	topics	string

	shutdownchan chan error

	subscription*	shell.PubSubSubscription
}

func (l* BaseListner) GetTopics() string {
	return l.topics
}

func (l* BaseListner) ThreadState() (s int){
	return l.threadstate
}

func (l* BaseListner) Shutdown() (err error) {

	if l.subscription != nil {

		l.threadstate = ListennerThread_Stop

		if err := l.subscription.Cancel(); err != nil {
			return nil
		}

	}

	return <- l.shutdownchan
}

func (l* BaseListner) StartListening() error {

	if l.state == nil {
		return errors.New("Befor StartListening must set this dapp state to listner instance.")
	}

	l.shutdownchan = make(chan error)

	go func( in chan error ){

		if subs, err := shell.NewLocalShell().PubSubSubscribe(l.topics); err == nil {

			log.Println("Listner : " + l.topics + " Start Listening.")
			l.threadstate = ListennerThread_Running

			l.subscription = subs

			for {

				msg, err := subs.Next()

				if err != nil {

					//异常前已经设置了线程状态为关闭，则为调用shutdown关闭，属于正常关闭，所以无错误的
					//否则是其他原因导致，则把状态设为Dead，表示意料之外的情况杀死了线程
					if l.threadstate == ListennerThread_Stop {
						in <- nil
						return
					} else {
						l.threadstate = ListennerThread_Dead
						in <- err
						return
					}

				} else {
					if l.threadstate == ListennerThread_Running {
						l.handleDelegate(msg)
					}
				}

			}

		} else {

			l.threadstate = ListennerThread_Dead
			in <- err

			return
		}

	}(l.shutdownchan)

	return nil
}