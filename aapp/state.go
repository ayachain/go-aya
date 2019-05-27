package aapp

type AAppStat int8

const (
	AAppStat_Unkown	AAppStat = 0
	AAppStat_Loaded		AAppStat = 1
	AAppStat_Daemon	AAppStat = 2
	AAppStat_Stoped		AAppStat = 3
)