package block

import (
	AStep "github.com/ayachain/go-aya/consensus/core/step"
	"github.com/ayachain/go-aya/consensus/impls/APOS/common/filter"
	"github.com/ipfs/go-ipfs/core"
)

func NewConsensusStep( ind *core.IpfsNode ) *AStep.AConsensusStep {

	sroot := AStep.NewStepRoot( "APOS_In_Block_Filter", ind, filter.MessageFilter)

	sroot.
		LinkNext("APOS_In_Block_Signature", SignaturerStep ).
		LinkNext("APOS_In_Block_Worker", WokerSwitcher)

	return sroot
}