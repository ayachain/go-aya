package block

type Block struct {

	/// block index
	Index int64 `json:"index"`

	/// prev block hash is a ipfs block CID
	Parent  string `json:"parent"`

	/// full block data CID
	ExtraData string `json:"extradata"`

	/// broadcasting time super master node package this block times.
	Timestamp int	`json:"timestamp"`

	/// chain id
	ChainID string `json:"chainid"`

}


/// only in create a new chain then use
type GenBlock struct {
	Block
	Award map[string]string `json:"award"`
	ConsensusRule string `json:"consensusrule"`
}

var (
	Genesis = &Block{Index: -4}
	Curr 	= &Block{Index: -3}
	Latest 	= &Block{Index: -2}
	Pending = &Block{Index: -1}
)