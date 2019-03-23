package response

const (
	GATEWAY_ERROR_Parser_Faild = -1000
	GATEWAY_ERROR_Verify_Faild = -1001
	GATEWAY_ERROR_Request_IO_Faild = -1002
	GATEWAY_ERROR_Publish_Tx_Faild = -1003
	GATEWAY_ERROR_OnlySupport_Post = -1004
	GATEWAY_ERROR_OnlySupport_Get = -1005
	GATEWAY_ERROR_MissParmas_TxHash = -1006
	GATEWAY_ERROR_MissParmas_DappNS = -1007
	GATEWAY_ERROR_Tx_NotFound = -1008
)

const (
	GATEWAY_SUCCESS	= 0
)


func CreateError( code int ) HttpResponseWriter {

	switch code {

	case GATEWAY_ERROR_Parser_Faild:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Analytical Failure of Transaction Parameters.", Body:nil }

	case GATEWAY_ERROR_Verify_Faild:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Signature verification failed.", Body:nil }

	case GATEWAY_ERROR_Request_IO_Faild:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Request read exception.", Body:nil }

	case GATEWAY_ERROR_OnlySupport_Post:
		return &HttpResponse{ HttpState:200, Code:code, Message:"The requested API only supports Post.", Body:nil }

	case GATEWAY_ERROR_OnlySupport_Get:
		return &HttpResponse{ HttpState:200, Code:code, Message:"The requested API only supports Get.", Body:nil }

	case GATEWAY_ERROR_MissParmas_TxHash:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Missing parmas 'txhash'", Body:nil }

	case GATEWAY_ERROR_MissParmas_DappNS:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Missing parmas 'dappns'", Body:nil }

	case GATEWAY_ERROR_Tx_NotFound:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Tx not found.", Body:nil }

	}

	return nil

}

func CreateSuccess( body interface{} ) HttpResponseWriter {

	return &HttpResponse{ HttpState:200, Code:GATEWAY_SUCCESS, Message:"Success", Body:body }

}