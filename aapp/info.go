package aapp

import "fmt"

type info struct {
	AAppns string
	Ver string
	VMVer string
	Copyright string
}

func ( inf *info ) GetChannelTopic() string {
	return fmt.Sprintf("%v_%v_%v", inf.AAppns, inf.Ver, inf.VMVer)
}