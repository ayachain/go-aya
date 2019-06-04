package receipt

import (
	EComm "github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
)

type aReceipt struct {
	ReceiptsAPI
	rawdb *leveldb.DB
}


func (r *aReceipt) GetTransactionReceipt( txhs EComm.Hash ) (*Receipt, error) {

	vbs, err := r.rawdb.Get( txhs.Bytes(), nil)

	if err != nil {
		return nil, err
	}

	rp := &Receipt{}

	err = rp.Decode( vbs )
	if err != nil {
		return nil, err
	}

	return rp, nil
}
