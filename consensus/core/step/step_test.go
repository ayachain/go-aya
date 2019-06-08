package step

import (
	"context"
	AWork "github.com/ayachain/go-aya/consensus/core/worker"
	Avdb "github.com/ayachain/go-aya/vdb"
	"github.com/ipfs/go-ipfs/core"
	"testing"
)


func TestInitialization(t *testing.T) {


	root := NewStepRoot(
		"Step1",
		nil,
		nil,
		 func(
		 	i interface{},
		 	cvfs *Avdb.CVFS,
		 	node *core.IpfsNode,
		 	group *AWork.TaskBatchGroup,
		 	) error {
			 println("Step1")
			 return nil
		 },
	)

	root.LinkNext("Step2", func(i interface{}, cvfs *Avdb.CVFS, node *core.IpfsNode, group *AWork.TaskBatchGroup) error {
		println("Step2")
		return nil
	}).LinkNext("Step3", func(i interface{}, cvfs *Avdb.CVFS, node *core.IpfsNode, group *AWork.TaskBatchGroup) error {
		println("Step3")
		return nil
	}).LinkNext("Step4",func(i interface{}, cvfs *Avdb.CVFS, node *core.IpfsNode, group *AWork.TaskBatchGroup) error {
		println("Step4")
		return nil
	})


	<- root.DoConsultation(context.TODO(), "Hello", nil)
	<- root.DoConsultation(context.TODO(), "Hello", nil)
	<- root.DoConsultation(context.TODO(), "Hello", nil)

	println("over")

}
