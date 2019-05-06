package module

import (
	"github.com/ayachain/go-aya/avm/miner/module/ipfs"
	"github.com/yuin/gopher-lua"
	LJson "layeh.com/gopher-json"
)

func InjectionAyaModules(l *lua.LState) {

	//config avm
	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
	} {
		if err := l.CallByParam(lua.P{
			Fn:      l.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			panic(err)
		}
	}

	ipfs.BasePathFunc = GetAvmBasePath

	l.PreloadModule("io", ipfs.Loader)
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