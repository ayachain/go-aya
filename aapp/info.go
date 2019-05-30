package aapp

type info struct {
	AAppns []string
	Ver string
	VMVer string
	Copyright string
}

func (inf info) GetChannelTopics() []string {

	var r []string

	for _, v := range inf.AAppns {

		r = append( r, "AAPP:" +  v + "_" + inf.Ver )

	}

	return r

}