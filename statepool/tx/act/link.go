package act

type LinkAct struct {
	BaseAct
}

func NewLinkAct(dpath string) BaseActInf{
	return &LinkAct{BaseAct{TStr:"LinkAct", DPath:dpath}}
}