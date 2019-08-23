package workflow

import (
	//"github.com/ayachain/go-aya/vdb"
	//ABlock "github.com/ayachain/go-aya/vdb/block"
)

const (
	OneDay 		= 3600 * 24
	OneMonth  	= OneDay * 30
	OneYear 	= OneMonth * 12
)

//func CanPos( block *AMBlock.MBlock, base vdb.CacheCVFS ) {
//
//	lblks, err := base.Blocks().GetBlocks(ABlock.BlockNameLatest)
//
//	if err != nil {
//		return
//	}
//
//	if len(lblks) <= 0 {
//		return
//	}
//
//	if lblks[0].Index < 11 {
//		return
//	}
//
//	pidx := base.Blocks().GetLatestPosBlockIndex()
//
//	pblks, err := base.Blocks().GetBlocks(pidx)
//
//	if err != nil {
//		return
//	}
//
//	if len(pblks) <= 0 {
//		return
//	}
//
//	if block.Timestamp - pblks[0].Timestamp < OneDay {
//		return
//	}
//
//	// get latest 11 block
//	var blk11Idx []uint64
//
//	for i := uint64(0); i < 11; i++ {
//		blk11Idx = append(blk11Idx, lblks[0].Index - i)
//	}
//
//	blks11, err := base.Blocks().GetBlocks( blk11Idx )
//	if err != nil {
//		return
//	}
//
//	var totalTimes uint64  = 0
//
//	for _, v := range blks11 {
//		totalTimes += v.Timestamp
//	}
//
//	if pblks[0].Timestamp - totalTimes / 11 > OneDay {
//
//		if DoPos( totalTimes / 11, base) == nil {
//			base.Blocks().SetLatestPosBlockIndex( block.Index )
//		}
//
//	}
//
//	return
//}
//
///// DayPos 2000000000000
///// EverMonth sub 3%
//func DoPos( ptime uint64, base vdb.CacheCVFS ) error {
//
//	return nil
//	//genblk, err := base.Blocks().GetBlocks(ABlock.BlockNameGen)
//	//
//	//if err != nil {
//	//	return err
//	//}
//	//
//	//distMonthCount := (ptime - genblk[0].Timestamp) % OneMonth
//	//
//	//prop := math.Pow( 0.97, float64(distMonthCount))
//	//
//	//posTotalAmount := uint64(int64(2000000000000 * prop))
//	//
//	//residueAmount := posTotalAmount
//	//
//	//// Super node pos profits
//	//superNds := base.Nodes().GetSuperNodeList()
//	//
//	//superNdBaseP := posTotalAmount / uint64(len(superNds))
//	//
//	//for _, snd := range superNds {
//	//
//	//	ast, err := base.Assetses().AssetsOf(snd.Owner)
//	//	if err != nil {
//	//		continue
//	//	}
//	//
//	//	ast.Avail += superNdBaseP
//	//
//	//	base.Assetses().Put( snd.Owner, ast )
//	//
//	//	residueAmount -= superNdBaseP
//	//}
//	//
//	//
//	//// ever nodes
//	//var nodelist[] *ANode.Node
//	//var totalLockedAmount uint64
//	//
//	//nit := base.Nodes().GetSnapshot().NewIterator(nil, nil )
//	//
//	//for nit.Next() {
//	//
//	//	nd := &ANode.Node{}
//	//
//	//	if err := nd.Decode(nit.Value()); err != nil {
//	//
//	//		nodelist = append(nodelist, nd)
//	//
//	//		totalLockedAmount += nd.Votes
//	//
//	//	}
//	//
//	//}
//	//
//	//for _, nd := range nodelist {
//	//
//	//	subndprofit := residueAmount / totalLockedAmount * nd.Votes
//	//
//	//	if subndprofit > 0 {
//	//
//	//		ast, err := base.Assetses().AssetsOf(nd.Owner)
//	//
//	//		if err != nil {
//	//			continue
//	//		}
//	//
//	//		ast.Avail += subndprofit
//	//		base.Assetses().Put( nd.Owner, ast )
//	//
//	//	}
//	//
//	//}
//	//
//	//return nil
//
//}


