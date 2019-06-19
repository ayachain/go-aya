package config

import (
	"encoding/json"
)

type config struct {
	IPFSRootPath string
	RepoKeyStorePath string
	WalletKeyStorePath string
}

var _config *config = nil

func Default() *config {

	if _config == nil {

		tWinConfig := `{"IPFSRootPath":"D:/AyaRopo/IPFS","RepoKeyStorePath":"D:/RepoKS","WalletKeyStorePath":"D:/WalletKS"}`
		//tMacConfig := `{"IPFSRootPath":"~/.ipfs,","RepoKeyStorePath ":"~/.ipfs/keystore","WalletKeyStorePath ":"~/.aya/keystore"}`
		_config = &config{}

		if err := json.Unmarshal([]byte(tWinConfig ), _config); err != nil {
			panic(err)
		}

	}

	return _config
}