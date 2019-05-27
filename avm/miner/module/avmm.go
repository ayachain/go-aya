package module

import (
	"github.com/ayachain/go-aya/avm/miner/module/ipfs"
	"github.com/yuin/gopher-lua"
	LJson "layeh.com/gopher-json"
)



/*
	luaLib{LoadLibName, OpenPackage},
	luaLib{BaseLibName, OpenBase},
	luaLib{TabLibName, OpenTable},
	luaLib{IoLibName, OpenIo},
	luaLib{OsLibName, OpenOs},
	luaLib{StringLibName, OpenString},
	luaLib{MathLibName, OpenMath},
	luaLib{DebugLibName, OpenDebug},
	luaLib{ChannelLibName, OpenChannel},
	luaLib{CoroutineLibName, OpenCoroutine},
*/

func InjectionAyaModules(l *lua.LState) {

	//config avm
	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
		//{lua.DebugLibName, lua.OpenDebug},
		//{lua.ChannelLibName, lua.OpenChannel},
	} {
		if err := l.CallByParam(lua.P{
			Fn:      l.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			panic(err)
		}
	}

	l.PreloadModule("io", ipfs.Loader)
	l.PreloadModule("json", LJson.Loader)
}