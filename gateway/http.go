package gateway

import (
	"github.com/ayachain/go-aya/gateway/aapp"
	"github.com/ayachain/go-aya/gateway/block"
	"github.com/ayachain/go-aya/gateway/tx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func DaemonHttpGateway() {

	go func() {

		echoServer := echo.New()
		//echoServer.Use(middleware.Logger())
		echoServer.Use(middleware.Recover())


		///AAPP ResutAPI
		echoServer.POST("/aapp/perfrom", aapp.PerfromHandle)
		echoServer.GET("/aapp/daemon",aapp.DaemonHandle)

		echoServer.GET("/tx/status", tx.TxStatusHandle)
		echoServer.GET("/block/get", block.BlockGetHandle)


		echoServer.Logger.Fatal(echoServer.Start("0.0.0.0:6001"))

	}()

}