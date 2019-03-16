package dappstate

const (
	DappListner_Broadcast 	= 1
	DappListner_TxCommit 	= 2
)

func CreateListener(ltype int, ds *DappState) Listener {

	var l Listener

	switch ltype {
	case DappListner_Broadcast:
		l =  NewBlockListner(ds)
	case DappListner_TxCommit:
		l = NewTxListner(ds)
	default:
		l = nil
	}

	return l
}