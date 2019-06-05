package step

import (
	"fmt"
)

type ConsensusChain struct {
	identifier string
	steps []ConsensusStep
}

func NewConsensusChain( idstr string, steps... ConsensusStep ) *ConsensusChain {

	if len(steps) <= 0 {
		panic(fmt.Errorf("ConsensusChain must hash one \"OverStep\" and one or more sub step"))
	}

	stepnames := map[string]ConsensusStep{}

	for _, v := range steps {
		if _, exist := stepnames[v.Identifier()]; exist {
			panic(fmt.Errorf("ConsensusStep identifier redefinition : %v", v.Identifier()))
		} else {
			stepnames[v.Identifier()] = v
		}
	}

	return &ConsensusChain{
		identifier:idstr,
		steps:steps,
	}

}

func (cc *ConsensusChain) GetStepRoot() ConsensusStep {
	return cc.steps[0]
}

func (cc *ConsensusChain) AppendSteps( step... ConsensusStep ) {

	cc.steps = append(cc.steps, step...)
	stepnames := map[string]ConsensusStep{}

	for _, v := range cc.steps {
		if _, exist := stepnames[v.Identifier()]; exist {
			panic(fmt.Errorf("ConsensusStep identifier redefinition : %v", v.Identifier()))
		} else {
			stepnames[v.Identifier()] = v
		}
	}

}

func (cc *ConsensusChain) LinkAllStep() {
	for i := 0; i < len(cc.steps) - 1; i++ {
		cc.steps[i].SetNextStep(cc.steps[i+1])
	}
}