package act

type TxRspAct struct {
	BaseAct
	BlockHash	string
	ResultState	string
}

func NewTxRspAct(dpath string, bhash string, shash string) (act* TxRspAct){
	return &TxRspAct{BaseAct{TStr:"TxRspAct",DPath:dpath}, bhash, shash}
}