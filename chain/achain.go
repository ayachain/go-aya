package chain

import (
	ABlock "github.com/ayachain/go-aya/block"
	"github.com/ipfs/go-cid"
)


type AyaChain interface {

	/// use GenBlock create a new block chain in aya
	GenChain( block *ABlock.GenBlock ) error

	/// put a new block and auto link prev
	PutBlock( block *ABlock.Block ) error

	/// get a existed block by index
	GetBlockByIndex( i int ) (*ABlock.Block, error)

	/// get block by cid, only return in chain block
	GetBlockByCID( cid cid.Cid ) (*ABlock.Block, error)

	/// get current confirmed block in this node
	GetCurBlock() (*ABlock.Block, error)

	/// use block create a ipfs-mfs
	GetVFS( block *ABlock.Block )

	/// verify chain of block index range start to end
	VerifyBlock( start, end int ) bool

	/// verify latest count block
	VerifyBlockByCount( bcount int ) bool

}

type aChain struct {

}