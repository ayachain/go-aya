package tx

import (
	RspFct "github.com/ayachain/go-aya/gateway/response"
	DSP "github.com/ayachain/go-aya/statepool"
	Atx "github.com/ayachain/go-aya/statepool/tx"
	"github.com/labstack/echo"
)

//Method:Get
//URL:http:127.0.0.1:6001/tx/status?dappns=QmbLSVGXVJ4dMxBNkxneThhAnXVBWdGp7i2S42bseXh2hS&txhash=0xa44f6b24f913428a66623490a87dffaa6d7ed28b5d2e4932b453cdf34343ecc7
func TxStatusHandle(c echo.Context) error {

	dappns := c.QueryParam("dappns")
	txhash := c.QueryParam("txhash")

	if len(dappns) <= 0 {
		return RspFct.CreateError(RspFct.GATEWAY_ERROR_MissParmas_DappNS).WriteToEchoContext(&c)
	}

	if len(txhash) <= 0 {
		return RspFct.CreateError(RspFct.GATEWAY_ERROR_MissParmas_TxHash).WriteToEchoContext(&c)
	}

	bindex,tx,s,rep := DSP.DappStatePool.GetTxStatus(dappns, txhash)

	switch s {
	default :
		return RspFct.CreateError(RspFct.GATEWAY_ERROR_Tx_NotFound).WriteToEchoContext(&c)

	case Atx.TxState_Pending:

		r := &Atx.TxReceipt{
			BlockIndex:bindex,
			Tx:*tx,
			TxHash:tx.GetSha256Hash(),
			Status:"Pending",
			Response:nil,
		}

		return RspFct.CreateSuccess(r).WriteToEchoContext(&c)


	case Atx.TxState_WaitPack:

		r := &Atx.TxReceipt{
			BlockIndex:bindex,
			Tx:*tx,
			TxHash:tx.GetSha256Hash(),
			Status:"WaitingPackage",
			Response:nil,
		}

		return RspFct.CreateSuccess(r).WriteToEchoContext(&c)

	case Atx.TxState_Confirm:
		return RspFct.CreateSuccess(rep).WriteToEchoContext(&c)
	}
}