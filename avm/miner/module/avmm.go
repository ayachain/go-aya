package module

import (
	"github.com/ayachain/go-aya/avm/miner/module/ipfs"
	"github.com/yuin/gopher-lua"
)

func InjectionAyaModules(l *lua.LState) {
	l.PreloadModule("ipfs", ipfs.Loader)
}