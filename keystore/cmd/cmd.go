package cmd

import (
	"errors"
	AKeyStore "github.com/ayachain/go-aya/keystore"
	"github.com/ethereum/go-ethereum/accounts"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"strings"
)

var (
	AKeyStoreServicesError = errors.New("aya keystore services expected")
	AKeyStoreServicesNotFoundError = errors.New("not found account address")
)

func findAccount( hexAddr string ) (accounts.Account, error) {

	aks := AKeyStore.ShareInstance()
	if aks == nil {
		return accounts.Account{}, AKeyStoreServicesError
	}

	for _, acc := range aks.Accounts() {
		if strings.EqualFold( acc.Address.String(), hexAddr ) {
			return acc, nil
		}
	}

	return accounts.Account{}, AKeyStoreServicesNotFoundError
}


var WalletCMDS = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "AyaChain Wallet Commands.",
	},
	Subcommands: map[string]*cmds.Command{
		"accounts" 			: 	accountsCmd,
		"new"					:	newAccountCmd,
		"delete"				:	deleteAccountCmd,
		"unlock"				:	unLockCmd,
		"update"				:	updateCmd,
		"sign"					:	signCmd,
	},

}