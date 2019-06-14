package impls

import (
	ACore "github.com/ayachain/go-aya/consensus/core"
	APos "github.com/ayachain/go-aya/consensus/impls/APOS"
	"github.com/ipfs/go-ipfs/core"
	"github.com/pkg/errors"
)

var (
	ErrNotSupportNotary		=		errors.New("not support consensus notary name")
)

//NewAPOSConsensusNotary( m vdb.CVFS, ind *core.IpfsNode )
func CreateNotary( cname string, ind *core.IpfsNode ) (ACore.Notary, error) {

	switch cname {
	case "APOS":
		return APos.NewAPOSConsensusNotary( ind ), nil

	default:
		return nil, ErrNotSupportNotary
	}

}