package block

import (
	RspFct "github.com/ayachain/go-aya/gateway/response"
	"github.com/ayachain/go-aya/statepool/tx"
	"github.com/ayachain/go-aya/utils"
	"github.com/labstack/echo"
	"strconv"
)

//http:127.0.0.1:6001/block/get?dappns=QmbLSVGXVJ4dMxBNkxneThhAnXVBWdGp7i2S42bseXh2hS&index=latest
func BlockGetHandle(c echo.Context) error {

	dappns := c.QueryParam("dappns")
	index := c.QueryParam("index")

	if len(dappns) <= 0 {
		return RspFct.CreateError(RspFct.GATEWAY_ERROR_MissParmas_DappNS).WriteToEchoContext(&c)
	}

	if len(index) <= 0 {
		return RspFct.CreateError(RspFct.GATEWAY_ERROR_MissParmas_Index).WriteToEchoContext(&c)
	}

	//index files stat
	ipath := "/" + dappns + "/_index/_bindex"
	ifstat, err := utils.AFMS_PathStat(ipath)

	if err != nil {
		return RspFct.CreateError(RspFct.GATEWAY_ERROR_MFSPathStat_ERROR).WriteToEchoContext(&c)
	}

	var blockHash []byte

	switch index {
	case "latest":

		blockHash, err = utils.AFMS_ReadFile(ipath, uint(ifstat.Size) - 46, 46)

		if err != nil {
			return RspFct.CreateError(RspFct.GATEWAY_ERROR_MFSIO_Expection).WriteToEchoContext(&c)
		}

	default:

		if bnub, err := strconv.Atoi(index); err != nil {
			return RspFct.CreateError(RspFct.GATEWAY_ERROR_Conversion_Error).WriteToEchoContext(&c)
		} else {

			blockHash, err = utils.AFMS_ReadFile(ipath, (uint(bnub) - 1) * 46, 46)

			if err != nil {
				return RspFct.CreateError(RspFct.GATEWAY_ERROR_MFSIO_Expection).WriteToEchoContext(&c)
			}
		}
	}

	if blk, err := tx.ReadBlock(string(blockHash)); err != nil {
		return RspFct.CreateError(RspFct.GATEWAY_ERROR_BlockRead_Expection).WriteToEchoContext(&c)
	} else {
		return RspFct.CreateSuccess(blk).WriteToEchoContext(&c)
	}

}