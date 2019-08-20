package msgcenter

type TrustedConfig struct {

	// Default 0.50 (50%)
	VotePercentage float32

	// 3
	SuperNodeMin uint

	// 3
	NodeTotalMin uint

}

var DefaultTrustedConfig = TrustedConfig{
	VotePercentage:0.5,
	SuperNodeMin:2,
	NodeTotalMin:2,
}