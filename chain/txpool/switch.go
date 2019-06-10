package txpool

import (
	"github.com/pkg/errors"

	AKeyStore "github.com/ayachain/go-aya/keystore"
	AMsgBlock "github.com/ayachain/go-aya/vdb/block"
	AMsgInfo "github.com/ayachain/go-aya/vdb/chaininfo"
	AMsgMBlock "github.com/ayachain/go-aya/vdb/mblock"
	AMsgMinied "github.com/ayachain/go-aya/vdb/minined"
	ATransaction "github.com/ayachain/go-aya/vdb/transaction"
)

func (pool *ATxPool) RawMessageSwitch( msg *AKeyStore.ASignedRawMsg ) error  {

	if !msg.Verify() {
		return ErrMessageVerifyExpected
	}

	switch msg.Content[0] {

	case AMsgBlock.MessagePrefix:

		//result := <- pool.notary.OnReceiveRawMessage(msg)
		//if result.Err != nil {
		//	// consensus failed
		//}
		//
		//block := AMsgBlock.Block{}
		//if err := block.RawMessageDecode( msg.Content ); err != nil {
		//	return err
		//}
		//
		//bcid, err:= cid.Decode(block.ExtraData)
		//if err != nil {
		//	return err
		//}
		//
		//if err := pool.UpdateBestBlock(bcid); err != nil {
		//	return err
		//}
		//
		//break


	case ATransaction.MessagePrefix:

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


	case AMsgMinied.MessagePrefix :

		pool.threadChans[AtxThreadsNameMining] <- msg

	default:

		return errors.New("undecode raw data")

	}

	return nil
}