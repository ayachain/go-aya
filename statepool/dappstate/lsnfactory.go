package dappstate

const (
	DappListner_Broadcast 	= 1
	DappListner_TxCommit 	= 2
)

func CreateListener(ltype int, ds *DappState) Listener {

	switch ltype {
	case DappListner_Broadcast:
		return NewBlockListner(ds)
	case DappListner_TxCommit:
		return NewTxListner(ds)
	}

	return nil
}