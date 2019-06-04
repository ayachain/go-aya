package signaturer

import (
	ACStep "github.com/ayachain/go-aya/consensus/core/step"
)

type SignatureAPI interface {

	ACStep.ConsensusStep

	//Recover( signature *RawSignature ) string
	//
	//SignRaw( content []byte, address string, passphrase ... string  ) (*RawSignature, error)
	//SignBlock( block *ABlock.Block, address string, passphrase ... string ) (*RawSignature, error)
	//SignTransaction( tx *ATx.Transaction, address string, passphrase ... string ) (*RawSignature, error)
	//SignHash( hash EComm.Hash, address string, passphrase ... string ) (*RawSignature, error)
	//
	//VerifyRaw( signture *RawSignature ) bool
	//VerifyToBlock( signture *RawSignature, address string ) ( *ABlock.Block, error )
	//VerifyToTransaction( signture *RawSignature, address string ) ( *ATx.Transaction, error )
	//VerifyToHash( signture *RawSignature, address string ) ( *EComm.Hash, error )

}