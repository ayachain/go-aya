package tx

import (
	"encoding/json"
	RspFct "github.com/ayachain/go-aya/gateway/response"
	DSP "github.com/ayachain/go-aya/statepool"
	Atx "github.com/ayachain/go-aya/statepool/tx"

	"io/ioutil"
	"net/http"
)

//http:127.0.0.1:6001/tx/perfrom
func TxPerfromHandle(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		RspFct.CreateError(RspFct.GATEWAY_ERROR_OnlySupport_Post).WriteToStream(&w);return
	}

	if bodybs, err := ioutil.ReadAll(r.Body); err != nil {

		RspFct.CreateError(RspFct.GATEWAY_ERROR_Request_IO_Faild).WriteToStream(&w);return

	} else {

		ptx := &Atx.Tx{}

		if err := json.Unmarshal(bodybs, ptx); err != nil {
			RspFct.CreateError(RspFct.GATEWAY_ERROR_Parser_Faild).WriteToStream(&w);return
		} else {

			if !ptx.VerifySign() {
				RspFct.CreateError(RspFct.GATEWAY_ERROR_Verify_Faild).WriteToStream(&w);return
			}

			if err := DSP.DappStatePool.PublishTx(ptx); err != nil {
				RspFct.CreateError(RspFct.GATEWAY_ERROR_Publish_Tx_Faild).WriteToStream(&w);return
			} else {
				RspFct.CreateSuccess(ptx.GetSha256Hash()).WriteToStream(&w);return
			}
		}
	}
}