package ChainStruct

import (
	"AyaChain/MemoryPool"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ipfs/go-ipfs-api"
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

	IpfsShell* 	shell.Shell

	peerType	string

	ayaChainTransactionPool_Commit_Subscription* shell.PubSubSubscription
}

func (tc* TransactionChain) Deamon() (err error) {

	tc.IpfsShell = shell.NewLocalShell()

	commitSubscription, err := tc.IpfsShell.PubSubSubscribe(WTopics_AyaChainTransactionPool_Commit)

	if err != nil {
		return err
	}

	tc.ayaChainTransactionPool_Commit_Subscription = commitSubscription

	go func() {

		for {

			switch tc.peerType {

			case AyaPeerType_Master:

				msg, err := tc.ayaChainTransactionPool_Commit_Subscription.Next()

				if err != nil {
					panic(err)
					continue
				}

				tx := &MemoryPool.Transaction{}

				if err := tx.Deserialize(msg.Data); err != nil {
					panic(err)
					continue
				}

				if v,err := tx.Verify(); !v && err != nil {
					panic(err)
					continue
				}

				//交易验证成功,加入交易池
				tc.GenerateNewBlock(hexutil.Encode(msg.Data))

				//开始广播最新的交易
				blockHex, err := tc.LatestBlock().EncodeToHex()

				if err != nil {
					panic(err)
					continue
				}

				if err = tc.IpfsShell.PubSubPublish(MTopics_AyaChainTransactionPool_Broadcast, blockHex); err != nil {
					panic(err)
					continue
				}

			case AyaPeerType_Worker:
			}
		}

	}()
}