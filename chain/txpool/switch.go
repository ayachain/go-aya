package txpool

import (
	"github.com/pkg/errors"

	AMsgBlock "github.com/ayachain/go-aya/chain/message/block"
	AMsgBReceipt "github.com/ayachain/go-aya/chain/message/blockreceipt"
	AMsgInfo "github.com/ayachain/go-aya/chain/message/chaininfo"
	AMsgMBlock "github.com/ayachain/go-aya/chain/message/miningblock"
	AMsgTx "github.com/ayachain/go-aya/chain/message/transaction"
	AKeyStore "github.com/ayachain/go-aya/keystore"
)

func (pool *ATxPool) rawMessageSwitch( msg *AKeyStore.ASignedRawMsg ) error  {

	if !msg.Verify() {
		return ErrMessageVerifyExpected
	}

	switch msg.Content[0] {

	case AMsgBlock.MessagePrefix:

		//If a confirmation block is received from the network, a notary is used to verify the
		// consensus link and write it in.
		result := <- pool.notary.OnReceiveRawMessage(msg)

		if result.Err != nil {
			// consensus failed
		}

		pool.UpdateBestBlock()

		break


	case AMsgTx.MessagePrefix:

		// If a transaction message is received from the network, it should be backed up locally.
		// Although currently running to the node may not be able to pack, in essence, the transaction
		// pool should be a network-wide synchronization to the queue.
		if err := pool.AddRawTransaction( msg ); err != nil {
			return err
		}

		break


	case AMsgMBlock.MessagePrefix:
		{

		}

	case AMsgInfo.MessagePrefix :
		{

		}

	case AMsgBReceipt.MessagePrefix :
		{

			var receipt = AMsgBReceipt.MsgRawMiningReceipt{}

			if err := receipt.Decode(msg.Content); err != nil {
				return err
			}

			addr, err := msg.ECRecover()
			if err != nil {
				return err
			}

			pool.AddConfrimReceipt( receipt.MBlockHash, receipt.RetCID, addr)

		}

	default:
		return errors.New("undecode raw data")
	}

	return nil
}