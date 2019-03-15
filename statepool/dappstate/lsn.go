package dappstate

import (
	"github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
)

const ListnerTopicPrefix  = "AyaChainListner."

const (
	ListennerThread_Stop 	= -1
	ListennerThread_Running = 0
	ListennerThread_Dead	= 2
)

type Listener interface {

	StartListening() error

	Shutdown() error

	Handle(msg *shell.Message)

	ThreadState() int

	GetTopics() string
}

type baseListner struct {

	Listener

	state*	DappState

	threadstate int

	topics	string

	shutdownchan chan error

	subscription*	shell.PubSubSubscription
}

func (l* baseListner) GetTopics() string {
	return l.topics
}

func (l* baseListner) ThreadState() (s int){
	return l.threadstate
}

func (l* baseListner) Shutdown() (err error) {

	if l.subscription != nil {

		l.threadstate = ListennerThread_Stop

		if err := l.subscription.Cancel(); err != nil {
			return nil
		}

	}

	return <- l.shutdownchan
}

func (l* baseListner) StartListening() error {

	if l.state == nil {
		return errors.New("Befor StartListening must set this dapp state to listner instance.")
	}

	l.shutdownchan = make(chan error)

	go func( in chan error ){

		if subs, err := shell.NewLocalShell().PubSubSubscribe(l.topics); err == nil {

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
						l.Handle(msg)
					}
				}

			}

		} else {

			l.threadstate = ListennerThread_Dead
			in <- err

			return
		}

	}(l.shutdownchan)
}