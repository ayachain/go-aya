package dappstate

func CreateListener(ltype int, ds *DappState) Listener {

	var l Listener

	switch ltype {
	case PubsubChannel_Block:
		l =  NewBlockListner(ds)
	case PubsubChannel_Tx:
		l = NewTxListner(ds)
	case PubsubChannel_Rsp:
		return NewRspListner(ds)
	default:
		l = nil
	}

	return l
}