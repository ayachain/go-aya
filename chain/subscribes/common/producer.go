package common

type Producer interface {

	Dowork() <- chan struct{}

	Shutdown() <- chan struct{}

}