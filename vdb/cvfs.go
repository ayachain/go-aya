package vdb

import (
	"context"
	"errors"
	AWrok "github.com/ayachain/go-aya/consensus/core/worker"
	AAssetses "github.com/ayachain/go-aya/vdb/assets"
	ABlock "github.com/ayachain/go-aya/vdb/block"
	AHeader "github.com/ayachain/go-aya/vdb/headers"
	AReceipts "github.com/ayachain/go-aya/vdb/receipt"
	ATx "github.com/ayachain/go-aya/vdb/transaction"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
)

var (
	aVFSDAGNodeConversionError = errors.New("conversion proto node expected")
)

type CVFS interface {
	Assetses() 					AAssetses.AssetsAPI		/// Asset
	Headers() 					AHeader.HeadersAPI		/// Indexes
	Blocks() 					ABlock.BlocksAPI		/// Body
	Transactions() 				ATx.TransactionAPI		/// Transaction
	Receipts() 					AReceipts.ReceiptsAPI	/// Receipt
	WriteBatchGroup( group *AWrok.TaskBatchGroup) error
	Close()
}

type aCVFS struct {
	CVFS
	*mfs.Root
	inode *core.IpfsNode
	ctx context.Context
	ctxCancel context.CancelFunc

	assetsServices AAssetses.AssetsAPI
	headerServices AHeader.HeadersAPI
	blockServices ABlock.BlocksAPI
	txServices ATx.TransactionAPI
	receiptServices AReceipts.ReceiptsAPI
}

//ctx context.Context, aappns string, pnode *dag.ProtoNode, ind *core.IpfsNode
func CreateVFS( baseBlock *ABlock.Block, ind *core.IpfsNode ) (CVFS, error) {

	vcid, err := cid.Cast( []byte(baseBlock.ExtraData) )
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel( context.Background() )
	root, err := newMFSRoot( ctx, vcid, ind )
	if err != nil {
		return nil, err
	}

	headerDir, err := AVdbComm.LookupDBPath( root,  AHeader.DBPATH )
	if err != nil {
		return nil, err
	}

	assetDir, err := AVdbComm.LookupDBPath(root, AAssetses.DBPATH)
	if err != nil {
		return nil, err
	}

	blockDir, err := AVdbComm.LookupDBPath(root, ABlock.DBPath)
	if err != nil {
		return nil, err
	}

	txsDir, err := AVdbComm.LookupDBPath(root, ATx.DBPath)
	if err != nil {
		return nil, err
	}

	receiptDir, err := AVdbComm.LookupDBPath(root, AReceipts.DBPath)
	if err != nil {
		return nil, err
	}

	var (
			headerServices	= AHeader.CreateServices( headerDir )
			assetsServices 	= AAssetses.CreateServices( assetDir )
			blockServices	= ABlock.CreateServices( blockDir, headerServices )
			txServices			= ATx.CreateServices( txsDir )
			receiptServices	= AReceipts.CreateServices(receiptDir)
	)

	vfs := &aCVFS{
		ctx:ctx,
		ctxCancel:cancel,
		Root:root,
		inode:ind,
		assetsServices 	: assetsServices,
		headerServices : headerServices,
		blockServices	: blockServices,
		txServices			: txServices,
		receiptServices : receiptServices,
	}

	return vfs, nil
}

func newMFSRoot( ctx context.Context, c cid.Cid, ind *core.IpfsNode ) ( *mfs.Root, error ) {

	rnd, err := ind.DAG.Get(ctx, c)
	if err != nil {
		return nil, err
	}

	pbnd, ok := rnd.(*merkledag.ProtoNode)
	if !ok {
		return nil, aVFSDAGNodeConversionError
	}

	mroot, err := mfs.NewRoot(ctx, ind.DAG, pbnd, nil)
	if err != nil {
		return nil, err
	}

	return mroot, nil
}

func ( vfs *aCVFS ) changeBlock( c cid.Cid ) error {

	var root *mfs.Root
	var err error

	root, err = newMFSRoot(vfs.ctx, c, vfs.inode)
	if err != nil {
		return err
	}

	if err = vfs.Root.Close(); err != nil {
		return err
	}

	vfs.Root = root

	defer func() {

		if err != nil && root != nil {
			root.Close()
		}

	}()

	return nil
}

func ( vfs *aCVFS ) Close() {

}

func ( vfs *aCVFS ) Assetses() AAssetses.AssetsAPI {
	return vfs.assetsServices
}

func ( vfs *aCVFS ) Headers() AHeader.HeadersAPI {
	return vfs.headerServices
}

func ( vfs *aCVFS ) Blocks() ABlock.BlocksAPI {
	return vfs.Blocks()
}

func ( vfs *aCVFS ) Transactions() ATx.TransactionAPI {
	return vfs.txServices
}

func ( vfs *aCVFS ) Receipts() AReceipts.ReceiptsAPI {
	return vfs.receiptServices
}

func ( vfs *aCVFS ) WriteBatchGroup( group *AWrok.TaskBatchGroup) error {
	return nil
}