package common

import (
	"github.com/ethereum/go-ethereum/accounts"
	EComm "github.com/ethereum/go-ethereum/common"
)

type RawMsgSignature interface {
	SignWithAccount( account accounts.Account ) error
}


type RawMsgVerify interface {
	ECRecove() (EComm.Address, error)
}