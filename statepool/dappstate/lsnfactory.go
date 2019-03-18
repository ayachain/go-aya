package dappstate

func CreateListener(ltype int, ds *DappState) Listener {

	var l Listener

	switch ltype {
	case PubsubChannel_Block:
		l =  NewBlockListner(ds)
	case PubsubChannel_Tx:
		l = NewTxListner(ds)
	default:
		l = nil
	}

	return l
}