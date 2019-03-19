package utils

import (
	"context"
	"encoding/json"
	"github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
	"io/ioutil"
	"strings"
)

const AFMS_FILE  	=	"file"
const AFMS_DIR  	=	"directory"

type AFMS_Stat struct {
	Hash			string
	Size			uint64
	CumulativeSize 	uint64
	ChildBlocks		uint64
	Type 			string
}

func (as *AFMS_Stat) IsDir() bool {
	return strings.EqualFold(as.Type, AFMS_DIR)
}

func (as *AFMS_Stat) IsFile() bool {
	return strings.EqualFold(as.Type, AFMS_FILE)
}

func AFMS_PathStat(mpath string) (stat *AFMS_Stat, err error) {

	if mpath[0] != '/' {
		return nil, errors.New("paths must start with a leading slash.")
	}

	if r, err := shell.NewLocalShell().Request("files/stat").Arguments(mpath).Send(context.Background()); err != nil {

		return nil, err

	} else {

		if r.Error != nil {

			return nil, errors.New(r.Error.Error())

		} else {

			if bs, err := ioutil.ReadAll(r.Output); err != nil {

				return nil, err

			} else {

				stat = &AFMS_Stat{}

				if json.Unmarshal(bs, stat) != nil {

					return nil, err

				} else {

					return stat,nil

				}

			}

		}
	}

}

//检测文件是否存在
func AFMS_IsPathExist(mpath string) bool {

	if _, err := AFMS_PathStat(mpath); err != nil {
		return false
	} else {
		return true
	}

}

func AFMS_RemovePath(mpath string) bool {
	return shell.NewLocalShell().Request("files/rm").Arguments(mpath).Option("r", true).Exec(context.Background(), nil) == nil
}

func AFMS_DownloadPathToDir(source string, dist string) bool {
	return shell.NewLocalShell().Request("files/cp").Arguments(source, dist).Exec(context.Background(), nil) == nil
}

//装载Dapp运行文件
func AFMS_ReloadDapp(nsp string, dpath string) bool {

	var mfsTpath string

	if nsp[0] == '/' {
		mfsTpath = nsp
	} else {
		mfsTpath = "/" + nsp
	}

	if AFMS_IsPathExist(mfsTpath) {
		//文件存在直接删除
		AFMS_RemovePath(mfsTpath)
	}

	return AFMS_DownloadPathToDir(dpath, mfsTpath)

}

func AFMS_DestoryDapp(nsp string) bool {

	var mfsTpath string

	if nsp[0] == '/' {
		mfsTpath = nsp
	} else {
		mfsTpath = "/" + nsp
	}

	return AFMS_RemovePath(mfsTpath)
}

func AFMS_ReadDappCode(path string) (code string, err error) {

	mfsTpath := "/" + path + "_source/main.lua"

	bs,err := shell.NewLocalShell().Request("files/read").Arguments(mfsTpath).Send(context.Background())

	if err != nil {
		return "", err
	} else {

		if bs.Error != nil {
			return "", errors.New( bs.Error.Error() )
		}

		if codebs,err := ioutil.ReadAll(bs.Output); err != nil {
			return "", err
		} else {
			return string(codebs), nil
		}
	}

}