package response

import (
	"encoding/json"
	cmds "github.com/ipfs/go-ipfs-cmds"
)

type Response struct {
	Code int
	Body interface{}
}

const SimpleSuccessBody = "Success"


func RawExpectedResponse( body interface{}, code... int ) []byte {

	var c int = -1

	if len(code) > 0 {
		c = code[0]
	}

	r := &Response{
		int(c),
		body,
	}

	bs, _ := json.Marshal(r)
	return bs
}


func RawSusccessResponse( body interface{}, code... uint ) []byte {

	var c uint = 0

	if len(code) > 0 {
		c = code[0]
	}

	r := &Response{
		int(c),
		body,
	}

	bs, _ := json.Marshal(r)
	return bs
}

func EmitSuccessResponse( re cmds.ResponseEmitter,  body interface{} ) error {

	r := &Response{
		0,
		body,
	}

	return cmds.EmitOnce(re, r)
}

func EmitErrorResponse( re cmds.ResponseEmitter, err error ) error {

	r := &Response{
		-1,
		err.Error(),
	}

	return cmds.EmitOnce(re, r)

}

