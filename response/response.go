package response

import cmds "github.com/ipfs/go-ipfs-cmds"

type Response struct {
	Code int
	Body interface{}
}

const SimpleSuccessBody = "Success"

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

