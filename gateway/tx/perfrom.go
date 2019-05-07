package tx

import (
	"encoding/json"
	RspFct "github.com/ayachain/go-aya/gateway/response"
	DSP "github.com/ayachain/go-aya/statepool"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	"github.com/labstack/echo"

	"io/ioutil"
)

//http:127.0.0.1:6001/tx/perfrom
func TxPerfromHandle(c echo.Context) error {

	if bodybs, err := ioutil.ReadAll(c.Request().Body); err != nil {

		return RspFct.CreateError(RspFct.GATEWAY_ERROR_Request_IO_Faild).WriteToEchoContext(&c)

	} else {

		ptx := &Atx.Tx{}

		if err := json.Unmarshal(bodybs, ptx); err != nil {
			return RspFct.CreateError(RspFct.GATEWAY_ERROR_Parser_Faild).WriteToEchoContext(&c)
		} else {

			if !ptx.VerifySign() {
				return RspFct.CreateError(RspFct.GATEWAY_ERROR_Verify_Faild).WriteToEchoContext(&c)
			}

			if err := DSP.DappStatePool.PublishTx(ptx); err != nil {
				return RspFct.CreateError(RspFct.GATEWAY_ERROR_Publish_Tx_Faild).WriteToEchoContext(&c)
			} else {
				return RspFct.CreateSuccess(ptx.GetSha256Hash()).WriteToEchoContext(&c)
			}
		}
	}
}