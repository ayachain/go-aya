package cvfs

import (
	"context"
	"errors"
	ABlock "github.com/ayachain/go-aya/block"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
)

var (
	aVFSDAGNodeConversionError = errors.New("conversion proto node expected")
)

type aCVFS struct {
	*mfs.Root
	inode *core.IpfsNode
	ctx context.Context
	ctxCancel context.CancelFunc
}

//ctx context.Context, aappns string, pnode *dag.ProtoNode, ind *core.IpfsNode
func CreateVFS( baseBlock *ABlock.Block, ind *core.IpfsNode ) (*aCVFS, error) {

	vcid, err := cid.Cast( []byte(baseBlock.ExtraData) )
	if err != nil {
		return nil, err
	}

	vfs := &aCVFS{ inode:ind }

	vfs.ctx, vfs.ctxCancel = context.WithCancel( context.Background() )

	root, err := newMFSRoot(vfs.ctx, vcid, ind)
	if err != nil {
		return nil, err
	}

	vfs.Root = root

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