package cmd

import cmds "github.com/ipfs/go-ipfs-cmds"

var WalletCMDS = &cmds.Command{

	Helptext:cmds.HelpText{
		Tagline: "AyaChain Wallet commands.",
	},
	Subcommands: map[string]*cmds.Command{
		"coinbase"	 		: coinBaseCmd,
		"unlock"			: unlockCMD,
		"lock"				: lockCMD,
		"transfer"			: transferCMD,
		"balanceOf"			: balanceOfCMD,
	},
}