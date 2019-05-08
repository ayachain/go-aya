package aapp

import (

	DSP "github.com/ayachain/go-aya/statepool"
	RspFct "github.com/ayachain/go-aya/gateway/response"

	"github.com/labstack/echo"
)

//装载指定的Dapp
//http:127.0.0.1:6001/aapp/Daemon?dappns=QmVUaqfbeW3qNmrZqAbNrAin5aikYzpVv6GRt82Un28pW8
func DaemonHandle(c echo.Context) error {

	dappNS := c.QueryParam("dappns")

	if len(dappNS) <= 0 {
		return RspFct.CreateError(RspFct.GATEWAY_ERROR_MissParmas_DappNS).WriteToEchoContext(&c)
	}

	if err := DSP.DappStatePool.AddDappStatDaemon(dappNS); err != nil {
		return err
	} else {
		return RspFct.CreateSuccess(RspFct.GATEWAY_SUCCESS).WriteToEchoContext(&c)
	}

}