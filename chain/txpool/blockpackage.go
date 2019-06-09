package txpool

import (
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
)

func (pool *ATxPool) blockPackageThread(ctx context.Context) {

	fmt.Println("ATxPool block package thread power on")
	defer fmt.Println("ATxPool block package thread power off")

	for {

		select {
		case <-ctx.Done():
			return

		default:

			pool.miningBlockLocker.Lock()
			defer pool.miningBlockLocker.Unlock()

			if pool.miningBlock != nil && len(pool.miningBlock.ExtraData) > 0 {

				if err := pool.DoBroadcast(pool.miningBlock); err != nil {
					break
				}

				ucide, err := cid.Decode(pool.miningBlock.ExtraData)
				if err != nil {
					break
				}

				if err := pool.UpdateBestBlock( ucide ); err != nil {
					break
				}

				pool.miningBlock = nil
			}

		}
	}

}