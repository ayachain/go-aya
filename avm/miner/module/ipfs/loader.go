package ipfs

import (
	"context"
	"errors"
	"github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-ipfs-files"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"strings"
)

var BasePathFunc func(l *lua.LState) string

func Loader(L *lua.LState) int {

	mod := L.SetFuncs(L.NewTable(), exports)

	L.Push(mod)

	return 1
}

var exports = map[string]lua.LGFunction{
	"write"	: files_write,
	"stat" 	: files_stat,
	"read" 	: files_read,
	"cp"	: files_cp,
	"exist" : files_exist,
	"create": files_create,
	"mkdir" : files_mkdir,
	"rm"	: files_rm,
	"ls"	: files_ls,
	"mv"	: files_mv,
}

func files_mv(l *lua.LState) int {

	json, err := ipfs_files_request_PathAndParmas(l, "files/mv")

	if err != nil {

		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))

	} else {

		l.Push(lua.LString(json))
		l.Push(lua.LNil)

	}

	return 2

}

func files_ls(l *lua.LState) int  {

	json, err := ipfs_files_request_PathAndParmas(l, "files/ls")

	if err != nil {

		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))

	} else {

		l.Push(lua.LString(json))
		l.Push(lua.LNil)

	}

	return 2

}

func files_mkdir(l *lua.LState) int  {

	json, err := ipfs_files_request_PathAndParmas(l, "files/mkdir")

	if err != nil {

		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))

	} else {

		l.Push(lua.LString(json))
		l.Push(lua.LNil)

	}

	return 2

}

func files_rm(l *lua.LState) int {

	json, err := ipfs_files_request_SourceDistParmas(l, "files/rm")

	if err != nil {

		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))

	} else {

		l.Push(lua.LString(json))
		l.Push(lua.LNil)

	}

	return 2
}

func files_create(l *lua.LState) int {

	path :=  l.CheckString(1)

	if !strings.HasPrefix(path,"/") {
		l.Push(lua.LNil)
		l.Push(lua.LString("Error: paths must start with a leading slash."))
	}

	path = BasePathFunc(l) + path

	shell := shell.NewLocalShell()

	fr := files.NewBytesFile([]byte(""))
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	reqb := shell.Request("files/write").Arguments(path).Body(fileReader).Option("e", true)

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

func files_exist(l *lua.LState) int {

	path := l.CheckString(1)

	if !strings.HasPrefix(path,"/") {
		l.Push(lua.LNil)
		l.Push(lua.LString("Error: paths must start with a leading slash."))
	}

	path = BasePathFunc(l) + path

	reqb := shell.NewLocalShell().Request("files/stat").Arguments(path)

	if response, err := reqb.Send(context.Background()); err != nil {

		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))
		return 2

	} else {

		if response.Error != nil {

			l.Push(lua.LBool(false))
			l.Push(lua.LNil)
			return 2

		} else {

			l.Push(lua.LBool(true))
			l.Push(lua.LNil)
			return 2

		}

	}

}

func files_cp(l *lua.LState) int{

	json, err := ipfs_files_request_SourceDistParmas(l, "files/cp")

	if err != nil {

		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))

	} else {

		l.Push(lua.LString(json))
		l.Push(lua.LNil)

	}

	return  2

}

func files_read(l *lua.LState) int {

	json, err := ipfs_files_request_PathAndParmas(l, "files/read")

	if err != nil {

		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))

	} else {

		l.Push(lua.LString(json))
		l.Push(lua.LNil)

	}

	return  2
}

func files_stat(l *lua.LState) int {

	json, err := ipfs_files_request_PathAndParmas(l, "files/stat")

	if err != nil {

		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))

	} else {

		l.Push(lua.LString(json))
		l.Push(lua.LNil)

	}

	return  2
}

func files_write(l *lua.LState) int {

	path :=  l.CheckString(1)
	data := l.CheckString(2)
	parmas := l.CheckTable(3)

	if !strings.HasPrefix(path,"/") {
		l.Push(lua.LNil)
		l.Push(lua.LString("Error: paths must start with a leading slash."))
	}

	path = BasePathFunc(l) + path

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

func ipfs_files_request_PathAndParmas(l *lua.LState, comm string) (res string, err error)  {

	var path string
	var parmas *lua.LTable

	if l.GetTop() >= 2 {
		path = l.CheckString(1)
		parmas = l.CheckTable(2)
	} else {
		path = l.CheckString(1)
	}

	if !strings.HasPrefix(path,"/") {
		l.Push(lua.LNil)
		l.Push(lua.LString("Error: paths must start with a leading slash."))
	}

	path = BasePathFunc(l) + path

	reqb := shell.NewLocalShell().Request(comm).Arguments(path)

	if parmas != nil {

		parmas.ForEach(func(k lua.LValue, v lua.LValue) {
			reqb.Option(lua.LVAsString(k), lua.LVAsString(v))
		})

	}


	if response, err := reqb.Send(context.Background()); err != nil {

		return "", err

	} else {

		if response.Error != nil {

			return "", errors.New(response.Error.Message)

		} else {

			if bs, err := ioutil.ReadAll(response.Output); err != nil {

				return "", err

			} else {

				return string(bs), nil

			}


		}

	}

}

func ipfs_files_request_SourceDistParmas(l *lua.LState, comm string) (res string, err error)  {

	var source string
	var dist string
	var parmas *lua.LTable

	if l.GetTop() >= 3 {

		source = l.CheckString(1)
		dist = l.CheckString(2)
		parmas = l.CheckTable(3)

	} else {

		source = l.CheckString(1)
		dist = l.CheckString(2)

	}


	if !strings.HasPrefix(source,"/ipfs") {
		l.Push(lua.LNil)
		l.Push(lua.LString("Error: paths must start with a leading slash."))
	}

	if !strings.HasPrefix(dist,"/") {
		l.Push(lua.LNil)
		l.Push(lua.LString("Error: paths must start with a leading slash."))
	}

	dist = BasePathFunc(l) + dist

	reqb := shell.NewLocalShell().Request(comm).Arguments(source,dist)

	parmas.ForEach(func(k lua.LValue, v lua.LValue) {
		reqb.Option(lua.LVAsString(k), lua.LVAsString(v))
	})

	if response, err := reqb.Send(context.Background()); err != nil {

		return "", err

	} else {

		if response.Error != nil {

			return "", errors.New(response.Error.Message)

		} else {

			if bs, err := ioutil.ReadAll(response.Output); err != nil {

				return "", err

			} else {

				return string(bs), nil

			}


		}

	}

}