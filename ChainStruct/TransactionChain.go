package ChainStruct

import (
	"github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
	"strings"
)

const (
	AyaPeerType_Master = "Master"
	AyaPeerType_Worker = "Worker"
)

//Worker Topics
const (
	WTopics_AyaChainTransactionPool_Commit = "AyaChainTransactionPool.Commit"
)

//Master Topics
const (
	MTopics_AyaChainTransactionPool_Broadcast = "AyaChainTransactionPool.Broadcast"
)


type TransactionChain struct {

	Chain
	sh* 			shell.Shell
	peerType		string
}

func (tc* TransactionChain) SetPeerType(t string) {
	tc.peerType = t
}

func (tc* TransactionChain) DeamonMasterPeer() (err error) {

	tc.sh = shell.NewLocalShell()

	if !strings.EqualFold(tc.peerType, AyaPeerType_Master) {
		return errors.New("This Peer is't MasterNode.")
	}

	if err := tc.master_ListenningTransactionCommit(); err != nil {
		return err
	}

	return nil
}

func (tc* TransactionChain) DeamonWorkerPeer() (err error) {

	tc.sh = shell.NewLocalShell()

	if !strings.EqualFold(tc.peerType, AyaPeerType_Worker) {
		return errors.New("This Peer is't WorkerNode.")
	}

	if err := tc.woker_ListenningTransactionBroadcast(); err != nil {
		return err
	}

	return nil
}


//关注其他节点提交的交易
func (tc* TransactionChain) master_ListenningTransactionCommit() (err error) {

	//关注提交交易的Topics
	if subscription, err := tc.sh.PubSubSubscribe(WTopics_AyaChainTransactionPool_Commit); err == nil {

		go func() {

			for {

				msg, err := subscription.Next()

				if err != nil {
					panic(err)
					continue
				}

				tx := &Transaction{}

				if err := tx.DecodeFromHex(string(msg.Data)); err != nil {
					panic(err)
					continue
				}

				if v,err := tx.Verify(); !v && err != nil {
					panic(err)
					continue
				}

				//交易验证成功,加入交易池
				tc.GenerateNewBlock(string(msg.Data))

				//开始广播最新的交易
				blockHex, err := tc.LatestBlock().EncodeToHex()

				if err != nil {
					panic(err)
					continue
				}

				if err = tc.sh.PubSubPublish(MTopics_AyaChainTransactionPool_Broadcast, blockHex); err != nil {
					panic(err)
					continue
				}
			}
		}()

	} else {
		return err
	}

	return nil
}


//关注主节点广播的已经确认的交易
func (tc* TransactionChain) woker_ListenningTransactionBroadcast() (err error) {

	if subscription, err := tc.sh.PubSubSubscribe(MTopics_AyaChainTransactionPool_Broadcast); err == nil {

		go func() {

			for {

				msg, err := subscription.Next()

				if err != nil {
					panic(err)
				}

				tcb := &Block{}

				if err = tcb.DecodeFromHex(string(msg.Data)); err == nil {
					tc.AppendBlock(tcb)
					tc.DumpPrint()
				}
			}
		}()
	} else {
		return err
	}

	return nil
}