package common

import "github.com/whyrusleeping/go-logging"

type AThreadRoleType string

const (
	ThreadRoleTypeConsumer AThreadRoleType = "Consumer"
	ThreadRoleTypeProcucer AThreadRoleType = "Producer"
)


type AThreadSemaphore string

const (
	ThreadSemaphoreStop AThreadSemaphore = "stop"
	ThreadSemaphoreConsumer AThreadSemaphore = "consumer"
	ThreadSemaphoreProducer AThreadSemaphore = "producer"
)


var log = logging.MustGetLogger("ATxPool")