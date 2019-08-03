package txpool

import (
	"context"
	"fmt"
)

func txPackageThread(ctx context.Context ) {

	fmt.Println("ATxPool Thread On: " + ATxPoolThreadTxPackage)

	pool := ctx.Value("Pool").(*ATxPool)

	pool.workingThreadWG.Add(1)

	pool.threadChans.Store(ATxPoolThreadTxPackage, make(chan []byte, ATxPoolThreadTxPackageBuff))

	defer func() {

		cc, exist := pool.threadChans.Load(ATxPoolThreadTxPackage)
		if exist {

			close( cc.(chan []byte) )

			pool.threadChans.Delete(ATxPoolThreadTxPackage)

		}

		pool.workingThreadWG.Done()

		fmt.Println("ATxPool Thread Off: " + ATxPoolThreadTxPackage)

	}()


	for {

		cc, _ := pool.threadChans.Load(ATxPoolThreadTxPackage)

		select {
		case <- ctx.Done():
			return

		case _, isOpen := <- cc.(chan []byte):

			if !isOpen {
				continue
			}

			if pool.miningBlock != nil {
				continue
			}

			/// if txpool queue pool is empty this method return nil
			mblk := pool.CreateMiningBlock()

			if mblk != nil {

				pool.miningBlock = mblk

				if err := pool.doBroadcast(mblk, pool.channelTopics[ATxPoolThreadMining]); err != nil {
					log.Error(err)
					return
				}

			}
		}

	}
}