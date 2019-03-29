package response

const (
	GATEWAY_ERROR_BlockRead_Expection = -996
	GATEWAY_ERROR_MFSIO_Expection = -997
	GATEWAY_ERROR_MFSPathStat_ERROR = -998
	GATEWAY_ERROR_Conversion_Error = -999
	GATEWAY_ERROR_Parser_Faild = -1000
	GATEWAY_ERROR_Verify_Faild = -1001
	GATEWAY_ERROR_Request_IO_Faild = -1002
	GATEWAY_ERROR_Publish_Tx_Faild = -1003
	GATEWAY_ERROR_OnlySupport_Post = -1004
	GATEWAY_ERROR_OnlySupport_Get = -1005
	GATEWAY_ERROR_MissParmas_TxHash = -1006
	GATEWAY_ERROR_MissParmas_DappNS = -1007
	GATEWAY_ERROR_MissParmas_Index = -1008
	GATEWAY_ERROR_Tx_NotFound = -1009

)

const (
	GATEWAY_SUCCESS	= 0
)


func CreateError( code int ) HttpResponseWriter {

	switch code {
	case GATEWAY_ERROR_MFSIO_Expection:
		return &HttpResponse{ HttpState:200, Code:code, Message:"MFS I/O exception.", Body:nil }

	case GATEWAY_ERROR_MFSPathStat_ERROR:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Read MFS path status exception.", Body:nil }

	case GATEWAY_ERROR_Conversion_Error:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Type conversion error.", Body:nil }

	case GATEWAY_ERROR_Parser_Faild:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Analytical failure of transaction parameters.", Body:nil }

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

	case GATEWAY_ERROR_MissParmas_Index:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Missing parmas 'index'", Body:nil }

	case GATEWAY_ERROR_Tx_NotFound:
		return &HttpResponse{ HttpState:200, Code:code, Message:"Tx not found.", Body:nil }

	}

	return nil

}

func CreateSuccess( body interface{} ) HttpResponseWriter {

	return &HttpResponse{ HttpState:200, Code:GATEWAY_SUCCESS, Message:"Success", Body:body }

}