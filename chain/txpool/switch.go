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

	case ATransaction.MessagePrefix:

		if err := pool.AddRawTransaction( msg ); err != nil {
			return err
		}

		break


	case AMsgBlock.MessagePrefix:
		pool.threadChans[AtxThreadsNameBlockPacage] <- msg

	case AMsgMinied.MessagePrefix:
		pool.threadChans[AtxThreadsNameReceiptListen] <- msg

	case AMsgMBlock.MessagePrefix:
		pool.threadChans[AtxThreadsNameMining] <- msg

	case AMsgInfo.MessagePrefix :



	default:

		return errors.New("undecode raw data")

	}

	return nil
}