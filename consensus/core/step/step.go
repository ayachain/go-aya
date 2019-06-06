package step

import (
	"context"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
)

type ConsensusStep interface {

	Identifier( ) string

	SetNextStep( s ConsensusStep )

	ChannelAccept() chan *ADog.MsgFromDogs

	NextStep() ConsensusStep

	Consensued( *ADog.MsgFromDogs )

	StartListenAccept( ctx context.Context )
}