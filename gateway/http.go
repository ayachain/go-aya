package gateway

import (
	"fmt"
	"github.com/ayachain/go-aya/gateway/tx"
	"net/http"
)

func DaemonHttpGateway() {

	go func() {

		http.HandleFunc("/tx/perfrom", tx.TxPerfromHandle)
		http.HandleFunc("/tx/status", tx.TxStatusHandle)

		if err := http.ListenAndServe("0.0.0.0:6001", nil); err != nil {
			panic(err)
		}

	}()

	fmt.Println("AyaGateWay API listening at: 0.0.0.0:6001")

}