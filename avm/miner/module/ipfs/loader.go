package ipfs

import (
	"context"
	"github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-ipfs-files"
	"github.com/yuin/gopher-lua"
)

func Loader(L *lua.LState) int {

	mod := L.SetFuncs(L.NewTable(), exports)

	L.Push(mod)

	return 1
}

var exports = map[string]lua.LGFunction{
	"write": ipfs_write,
}

func ipfs_write(l *lua.LState) int {

	path := l.CheckString(1)
	data := l.CheckString(2)
	parmas := l.CheckTable(3)

	shell := shell.NewLocalShell()

	fr := files.NewBytesFile([]byte(data))
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	reqb := shell.Request("files/write").Arguments(path).Body(fileReader)

	parmas.ForEach(func(k lua.LValue, v lua.LValue) {
		reqb.Option(lua.LVAsString(k), lua.LVAsString(v))
	})

	err := reqb.Exec(context.Background(), nil)

	if err != nil {
		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))
	} else {
		l.Push(lua.LNil)
		l.Push(lua.LNil)
	}

	return 2
}