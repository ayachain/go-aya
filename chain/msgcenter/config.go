package msgcenter

type TrustedConfig struct {

	VotePercentage float32

	SuperNodeMin uint

	NodeTotalMin uint

}

var DefaultTrustedConfig = TrustedConfig{
	VotePercentage:0.5,
	SuperNodeMin:2,
	NodeTotalMin:2,
}