package dappstate

const (
	DappListner_Broadcast = 1
)

func CreateListener(ltype int, ds *DappState) Listener {

	switch ltype {
	case DappListner_Broadcast:
		return NewBlockListner(ds)
	}

	return nil
}