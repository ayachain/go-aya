package txpool

import (
	"context"
	AKeyStore "github.com/ayachain/go-aya/keystore"
)

/// Super Node Mode Specialization
func (pool *ATxPool) threadBlockPackage(pctx context.Context) <- chan error {

	replay := make(chan error)

	go func() {

		for {

			select {
			case <- pctx.Done():
				break

			case ret := <- pool.doPackingBlock:

				confblok := pool.miningBlock.ConfirmByCid(ret)

				srawmsg, err := AKeyStore.CreateMsg(confblok.Encode(), pool.ownerAccount)
				if err != nil {
					replay <- err
					break
				}

				if err := pool.DoBroadcast( srawmsg ); err != nil {
					replay <- err
					break
				}

				replay <- nil

			}
		}
	}()

	return replay
}