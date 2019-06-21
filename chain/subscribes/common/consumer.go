package common

type Consumer interface {

	Dowork() <- chan struct{}

	Shutdown() <- chan struct{}

}