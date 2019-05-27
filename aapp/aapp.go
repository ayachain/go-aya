package aapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	iCore "github.com/ayachain/go-aya/ipfsapi"
	iFiles "github.com/ipfs/go-ipfs-files"
	"io"
	"io/ioutil"
	"time"
)

type AAppPath string

const (
	AAppPath_Script 			AAppPath = "/Script"
	AAppPath_Resources 	AAppPath = "/Resource"
	AAppPath_Index 			AAppPath = "/Index"
	AAppPath_Evn 				AAppPath = "/Evn"
)

func (e AAppPath) ToString() string {
	switch (e) {
	case AAppPath_Script			: return "Script"
	case AAppPath_Resources		: return "Resource"
	case AAppPath_Index			: return "Index"
	case AAppPath_Evn				: return "Evn"
	default: return "UNKNOWN"
	}
}

type aapp struct {
	State AAppStat
	Info info
	CreateTime int64
	SubDir map[AAppPath] iFiles.Directory
}

func NewAApp( aappns string ) ( ap *aapp, err error ) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	path, err := iCore.IAPI.Name().Resolve(ctx, aappns)
	if err != nil {
		return nil, errors.New( fmt.Sprintf(`NewAApp：解析IPNS路径失败,"%v"`, err.Error() ) )
	}

	if fn, err := iCore.IAPI.Unixfs().Get(ctx, path); err != nil {

		return nil, errors.New( fmt.Sprintf(`NewAApp：IPFS路径读取失败,"%v"`, err.Error() ) )

	} else {

		var aappdir iFiles.Directory
		switch f := fn.(type) {
		case iFiles.File:
			return nil, errors.New(fmt.Sprintf(`NewAApp：AApps "%v" 对应了一个文件，AAppns应当对应目录。`))
		case iFiles.Directory:
			aappdir = f
			break
		default:
			return nil, errors.New(fmt.Sprintf(`NewAApp：AApps "%v" 不是一个正确的路径。`))
		}

		ap = &aapp{
			State:AAppStat_Unkown,
			CreateTime:time.Now().Unix(),
		}

		ap.SubDir[AAppPath_Script] = findSubDir(aappdir, AAppPath_Script.ToString())
		ap.SubDir[AAppPath_Resources] = findSubDir(aappdir, AAppPath_Resources.ToString())
		ap.SubDir[AAppPath_Index] = findSubDir(aappdir, AAppPath_Index.ToString())
		ap.SubDir[AAppPath_Evn] = findSubDir(aappdir, AAppPath_Evn.ToString())

		for k, v := range ap.SubDir {
			if v == nil {
				return nil, errors.New(fmt.Sprintf(`NewAApp："%v"对应的AApp中，缺少目录"%v"`, aappns, k))
			}
		}

		if sfn := findSubFile( ap.SubDir[AAppPath_Evn], "info" ); sfn == nil {
			return nil, errors.New(fmt.Sprintf(`NewAApp："%v"目录中未找到info文件`, aappns))
		} else {

			if bs, err := ioutil.ReadAll(sfn.(io.Reader)); err != nil {
				return nil, errors.New(`NewAApp：info文件读取失败`)
			} else {
				if err := json.Unmarshal(bs, ap.Info); err != nil {
					return nil, errors.New(`NewAApp：info文件解析失败`)
				}
			}

		}

		return ap, nil
	}

}

func findSubFile(  root iFiles.Directory, nodename string ) iFiles.File {

	nd := findSubNode(root, nodename)
	switch nd.(type) {
	case iFiles.File:
		return nd.(iFiles.File)
	default:
		return nil
	}
}

func findSubDir(  root iFiles.Directory, nodename string ) iFiles.Directory {
	nd := findSubNode(root, nodename)
	switch nd.(type) {
	case iFiles.Directory:
		return nd.(iFiles.Directory)
	default:
		return nil
	}
}

func findSubNode( root iFiles.Directory, nodename string ) iFiles.Node {

	it := root.Entries()

	for it.Name() != nodename {
		if ! it.Next() {
			return nil
		}
	}

	return it.Node()
}