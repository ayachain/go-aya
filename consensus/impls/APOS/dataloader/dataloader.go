// "Data Loader" is used to prepare data that needs to be downloaded beforehand,
// which mainly includes IPFS requests for the required data, because the data
// is all read-only, downloading the required data beforehand and then entering
// the next stage.

package dataloader

import (
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
	APosComm "github.com/ayachain/go-aya/consensus/impls/APOS/common"
	"github.com/ipfs/go-ipfs/core"
)


type DataLoader struct {

	nextStep ACStep.ConsensusStep

	acceptChan chan *ADog.MsgFromDogs

	identifier string

	ind *core.IpfsNode

}

func NewDataLoader(ind *core.IpfsNode) *DataLoader {
	return &DataLoader{
		identifier : "APOS-Step-2-DataLoader",
		acceptChan : make(chan *ADog.MsgFromDogs, APosComm.StepDataLoaderChanSize),
	}
}