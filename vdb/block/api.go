package block

type BlocksAPI interface {

	DBKey()	string

	GetBlocks ( iorc... interface{} ) ([]*Block, error)
}