package property

import (
	"fmt"
	"github.com/ipfs/go-cid"
	"testing"
)

func TestEncodeAndDecode(t *testing.T) {

	tcid, err := cid.Decode("Qma8HphKVArNubmSKkFGfTD1HPyBTFRx8acj1CNySAzrNe")
	if err != nil {
		t.Fatal(err)
	}

	srecord := &Record{
		Version:byte(1),
		Avail: 1000,
		Vote: 1000,
		ExtraCid :tcid,
	}

	ecd := srecord.Encode()
	fmt.Println( ecd )

	r2 := new( Record )
	if err := r2.Decode(ecd); err != nil {
		t.Fatal(err)
	}

	fmt.Println( r2.Version )
	fmt.Println( r2.Avail )
	fmt.Println( r2.Vote )
	fmt.Println( r2.ExtraCid.String() )
}