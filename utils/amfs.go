package utils

import (
	"context"
	"encoding/json"
	"github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-ipfs-files"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"strings"
)

//AFMS的ipfs files 相关子命令全部采用 -f=false 调用，若需要刷新到磁盘，需要手动调用flush，或者使用AFMS_FlushPath进行刷写磁盘
//禁用自动同步，因为每次都是计算一个块中到所有交易，每次修改都同步到磁盘，会产生大量Disk I/O 降低效率
//调用AFMS实现方法绝大多是情况都工作在内存中，是一些不确定都，或者尚未计算完全的交易

const AFMS_FILE  	=	"file"
const AFMS_DIR  	=	"directory"

type AFMS_Stat struct {
	Hash			string
	Size			int
	CumulativeSize 	int
	ChildBlocks		int
	Type 			string
}

type AFMS_Entrie struct {
	Name string
	Type int
	Size int
	Hash string
}

type AFMS_ListResponse struct {
	Entries	[]AFMS_Entrie
}

func (as *AFMS_Stat) IsDir() bool {
	return strings.EqualFold(as.Type, AFMS_DIR)
}

func (as *AFMS_Stat) IsFile() bool {
	return strings.EqualFold(as.Type, AFMS_FILE)
}

func AFMS_PathStat(mpath string) (stat *AFMS_Stat, err error) {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	if r, err := shell.NewLocalShell().Request("files/stat").Arguments(mpath).Option("flush",false).Send(context.Background()); err != nil {

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

func AFMS_ReplaceFile(mpath string, data []byte) error {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	stat, err := AFMS_PathStat(mpath)

	if err != nil {
		return err
	} else if stat.Type != AFMS_FILE {
		return errors.New("AFMS_ReplaceFile : " + mpath + " not a file.")
	}

	fr := files.NewBytesFile(data)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	reqb := shell.NewLocalShell().Request("files/write").Arguments(mpath).Body(fileReader).Option("t",true).Option("flush",false)

	if err := reqb.Exec(context.Background(), nil); err != nil {
		return err
	} else {
		return nil
	}


}

func AFMS_GetReader(mpath string) (io.ReadCloser, error) {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	req := shell.NewLocalShell().Request("files/read").Arguments(mpath)

	if bs, err := req.Send(context.Background()); err != nil {
		return nil, err
	} else {
		return bs.Output,nil
	}

}

func AFMS_IsPathExist(mpath string) bool {

	if _, err := AFMS_PathStat(mpath); err != nil {
		return false
	} else {
		return true
	}

}

func AFMS_RemovePath(mpath string) bool {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	return shell.NewLocalShell().Request("files/rm").Arguments(mpath).Option("flush",false).Option("r", true).Exec(context.Background(), nil) == nil
}

func AFMS_ReadScript(mpath string) (content string, err error) {
	if bs, err := AFMS_ReadFile(mpath,0,0); err != nil {
		return "", err
	} else {
		return string(bs), nil
	}
}

func AFMS_ReadFile(mpath string, offset uint, size uint) (content []byte, err error) {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	req := shell.NewLocalShell().Request("files/read").Arguments(mpath)

	if size > 0 {
		req.Option("o", offset).Option("n", size)
	}

	bs,err := req.Send(context.Background())

	if err != nil {
		return nil, err
	} else {

		if bs.Error != nil {
			return nil, errors.New( bs.Error.Error() )
		}

		if c,err := ioutil.ReadAll(bs.Output); err != nil {
			return nil, err
		} else {
			return c, nil
		}
	}

}

func AFMS_ReadDappCode(dappns string) (code string, err error) {

	mfsTpath := "/" + dappns + "/_dapp/main.lua"

	bs,err := shell.NewLocalShell().Request("files/read").Arguments(mfsTpath).Option("flush",false).Send(context.Background())

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

func AFMS_RenameFile(bpath string, s string, d string) error {

	if !strings.HasPrefix(bpath,"/") {
		bpath = "/" + bpath
	}

	source := bpath + "/" + s
	dist := bpath + "/" + d

	return shell.NewLocalShell().Request("files/mv").Arguments(source,dist).Option("flush",false).Exec(context.Background(), nil)
}

func AFMS_CreateFile(mpath string, data[] byte) error {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	shell := shell.NewLocalShell()

	fr := files.NewBytesFile([]byte(data))
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	reqb := shell.Request("files/write").Arguments(mpath).Body(fileReader).Option("e",true).Option("p",true).Option("flush",false)

	if err := reqb.Exec(context.Background(), nil); err != nil {
		return err
	} else {
		return nil
	}
}

func AFMS_WriteFile(mpath string, seek int, data[] byte) error {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	stat, err := AFMS_PathStat(mpath)

	if err != nil {
		return err
	} else if stat.Type != AFMS_FILE {
		return errors.New("AFMS_FileAppend : " + mpath + " not a file.")
	}

	fr := files.NewBytesFile(data)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	reqb := shell.NewLocalShell().Request("files/write").
		Arguments(mpath).
		Body(fileReader).
		Option("o",seek).
		Option("flush",false)

	if err := reqb.Exec(context.Background(), nil); err != nil {
		return err
	} else {
		return nil
	}

}

func AFMS_FileAppend(mpath string, data[] byte) error {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	stat, err := AFMS_PathStat(mpath)

	if err != nil {
		return err
	} else if stat.Type != AFMS_FILE {
		return errors.New("AFMS_FileAppend : " + mpath + " not a file.")
	}

	fr := files.NewBytesFile(data)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	reqb := shell.NewLocalShell().Request("files/write").
		Arguments(mpath).
		Body(fileReader).
		Option("o",stat.Size).
		Option("flush",false)

	if err := reqb.Exec(context.Background(), nil); err != nil {
		return err
	} else {
		return nil
	}

}

func AFMS_Mkdir(mpath string, parent bool) error {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	return shell.NewLocalShell().Request("files/mkdir").Arguments(mpath).Option("p", parent).Exec(context.Background(), nil);
}

func AFMS_FlushPath(mpath string) error {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	return shell.NewLocalShell().Request("files/flush").Arguments(mpath).Exec(context.Background(),nil)
}

func AFMS_PathLists(mpath string) (es []AFMS_Entrie, err error) {

	if !strings.HasPrefix(mpath,"/") {
		mpath = "/" + mpath
	}

	bs, err := shell.NewLocalShell().Request("files/ls").Arguments(mpath).Option("U", true).Send(context.Background())

	if err != nil {
		return nil, err
	} else {

		if bs.Error != nil {
			return nil, errors.New( bs.Error.Error() )
		}

		if c,err := ioutil.ReadAll(bs.Output); err != nil {
			return nil, err
		} else {

			l := &AFMS_ListResponse{}
			json.Unmarshal(c, l)

			return l.Entries, nil
		}
	}

}

func AFMS_DownloadPathToDir(ipfshash string, dist string) bool {

	ipath, err := resolveIPFSRef(ipfshash)

	if err != nil {
		return false
	}

	if !strings.HasPrefix(dist,"/") {
		dist = "/" + dist
	}

	return shell.NewLocalShell().Request("files/cp").Arguments(ipath, dist).Option("flush",false).Exec(context.Background(), nil) == nil
}

func AFMS_ReloadDapp(ipfshash string, mfspath string) bool {

	ipath, err := resolveIPFSRef(ipfshash)

	if err != nil {
		return false
	}

	if !strings.HasPrefix(mfspath, "/") {
		mfspath = "/" + mfspath
	}

	//查看本地MFS中是否已经存在此DAPP的缓存数据
	originStat, err := AFMS_PathStat(mfspath)

	if err != nil{

		//不存在则下载
		return AFMS_DownloadPathToDir(ipath, mfspath)

	} else {

		//存在并且Hash相同则说明不需要修改，负责删除后重新下载
		if originStat.Hash == ipath[6:] {

			return true

		} else {

			AFMS_RemovePath(mfspath)
			return AFMS_DownloadPathToDir(ipath, mfspath)

		}

	}
}

func resolveIPFSRef(ipfshash string) (string, error) {

	var ipath string
	var err error

	if strings.HasPrefix(ipfshash, "/ipfs") {

		//ipfs路径
		ipath = ipfshash

	} else if strings.HasPrefix(ipfshash, "/ipns") {

		//ipns路径需要resolve
		ipath, err = shell.NewLocalShell().ResolvePath(ipfshash)

		if err != nil {
			return "", err
		}

	} else {

		//只有路径，先尝试进行resolve，若不成功则当作ipfs路径处理
		ipath, err = shell.NewLocalShell().ResolvePath( "/ipns/" + ipfshash)

		if err != nil {
			ipath = "/ipfs/" + ipfshash
		} else {
			ipath = "/ipfs/" + ipath
		}
	}

	return ipath, nil
}