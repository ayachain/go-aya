package common

type AThreadRoleType string

const (
	ThreadRoleTypeConsumer AThreadRoleType = "Consumer"
	ThreadRoleTypeProcucer AThreadRoleType = "Producer"
)


type AThreadSemaphore string

const (
	ThreadSemaphoreStop AThreadSemaphore = "stop"
	ThreadSemaphoreSleep AThreadSemaphore = "sleep"
	ThreadSemaphoreResume AThreadSemaphore = "resume"
)