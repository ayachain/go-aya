package block

import (
	RspFct "github.com/ayachain/go-aya/gateway/response"
	"github.com/ayachain/go-aya/statepool/tx"
	"github.com/ayachain/go-aya/utils"
	"net/http"
	"strconv"
)

//http:127.0.0.1:6001/block/get?dappns=QmbLSVGXVJ4dMxBNkxneThhAnXVBWdGp7i2S42bseXh2hS&index=latest
func BlockGetHandle(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		RspFct.CreateError(RspFct.GATEWAY_ERROR_OnlySupport_Get).WriteToStream(&w);return
	}

	dappns := r.URL.Query().Get("dappns")

	if len(dappns) <= 0 {
		RspFct.CreateError(RspFct.GATEWAY_ERROR_MissParmas_DappNS).WriteToStream(&w);return
	}


	index := r.URL.Query().Get("index")

	if len(index) <= 0 {
		RspFct.CreateError(RspFct.GATEWAY_ERROR_MissParmas_Index).WriteToStream(&w);return
	}

	//index files stat
	ipath := "/" + dappns + "/_index/_bindex"
	ifstat, err := utils.AFMS_PathStat(ipath)

	if err != nil {
		RspFct.CreateError(RspFct.GATEWAY_ERROR_MFSPathStat_ERROR).WriteToStream(&w);return
	}

	var blockHash []byte

	switch index {
	case "latest":
		blockHash, err = utils.AFMS_ReadFile(ipath, ifstat.Size - 46, 46)

		if err != nil {
			RspFct.CreateError(RspFct.GATEWAY_ERROR_MFSIO_Expection).WriteToStream(&w);return
		}

	default:

		if bnub, err := strconv.Atoi(index); err != nil {
			RspFct.CreateError(RspFct.GATEWAY_ERROR_Conversion_Error).WriteToStream(&w);return
		} else {
			blockHash, err = utils.AFMS_ReadFile(ipath, (bnub - 1) * 46, 46)

			if err != nil {
				RspFct.CreateError(RspFct.GATEWAY_ERROR_MFSIO_Expection).WriteToStream(&w);return
			}
		}
	}

	if blk, err := tx.ReadBlock(string(blockHash)); err != nil {
		RspFct.CreateError(RspFct.GATEWAY_ERROR_BlockRead_Expection).WriteToStream(&w);return
	} else {
		RspFct.CreateSuccess(blk).WriteToStream(&w);return
	}

}