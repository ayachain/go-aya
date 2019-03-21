package module

import (
	"github.com/ayachain/go-aya/avm/miner/module/ipfs"
	"github.com/yuin/gopher-lua"
	LJson "layeh.com/gopher-json"
)

func InjectionAyaModules(l *lua.LState) {

	ipfs.BasePathFunc = GetAvmBasePath

	l.PreloadModule("ipfs", ipfs.Loader)
	l.PreloadModule("json", LJson.Loader)

}

var avmBasePathMrg map[*lua.LState]string

func SetAvmBasePath(l *lua.LState, path string) {

	if avmBasePathMrg == nil {
		avmBasePathMrg = map[*lua.LState]string{}
	}

	avmBasePathMrg[l] = path

}

func GetAvmBasePath(l *lua.LState) string {

	p, exist := avmBasePathMrg[l]

	if !exist {
		return ""
	} else {
		return p
	}

}