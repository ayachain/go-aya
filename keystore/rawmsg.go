package keystore

import (
	"encoding/json"
	"github.com/ayachain/go-aya/vdb/common"
	"github.com/ethereum/go-ethereum/accounts"
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type ASignedRawMsg struct {
	From  	EComm.Address 	`json:"F"`
	Content []byte			`json:"C"`
	Sig		[]byte			`json:"S"`
}

func BytesToRawMsg( rawbs []byte ) (*ASignedRawMsg, error) {

	r := &ASignedRawMsg{}

	if err := json.Unmarshal(rawbs, r); err != nil {
		return nil, err
	}

	return r, nil

}

func CreateMsgBySig( content []byte, sig []byte ) *ASignedRawMsg {

	return &ASignedRawMsg{
		Content:content,
		Sig:sig,
	}

}

func CreateMsgByRawCoder ( coder common.RawDBCoder, account accounts.Account ) (*ASignedRawMsg, error) {
	return CreateMsg(coder.Encode(), account)
}

func CreateMsg( content []byte, account accounts.Account ) (*ASignedRawMsg, error) {

	hs := crypto.Keccak256Hash(content)

	sig, err := ShareInstance().SignHash(account, hs.Bytes())
	if err != nil {
		return nil, err
	}

	return &ASignedRawMsg{
		From:account.Address,
		Content:content,
		Sig:sig,
	}, nil

}

func (raw *ASignedRawMsg) Verify() bool {

	addr, err := raw.ECRecover()
	if err != nil {
		return false
	}

	return *addr == raw.From
}

func (raw *ASignedRawMsg) ECRecover() (*EComm.Address, error){

	hs := crypto.Keccak256Hash(raw.Content)

	pubkey, err := crypto.SigToPub( hs.Bytes(), raw.Sig )
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(*pubkey)

	return &addr, nil
}

func (raw *ASignedRawMsg) Bytes() ([]byte, error) {
	return json.Marshal(raw)
}