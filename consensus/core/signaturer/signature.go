package signaturer

type RawSignature struct {
	P		[64]byte
	M		[2]byte
	S		[64]byte
	Content []byte
}