package block

type BlocksAPI interface {
	GetBlocks ( iorc... interface{} ) ([]*Block, error)
}