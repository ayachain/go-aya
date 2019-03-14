package act

type PerfromAct struct {
	BaseAct
	Method		string
	Parmas[]	string
}

func NewPerfromAct(dpath string, method string, parmas[] string) (act* PerfromAct){
	return &PerfromAct{BaseAct{"PerfromAct", dpath}, method, parmas}
}