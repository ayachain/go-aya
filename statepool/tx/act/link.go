package act

type LinkAct struct {
	BaseAct
}

func NewLinkAct(dpath string) (act* LinkAct){
	return &LinkAct{BaseAct{"LinkAct", dpath}}
}