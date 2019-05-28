package aapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ipfs/go-ipfs/core"
	dag "github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	"github.com/ipfs/interface-go-ipfs-core"
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
	VMFS *mfs.Root
}

func NewAApp( aappns string, api iface.CoreAPI, ind *core.IpfsNode ) ( ap *aapp, err error ) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	path, err := api.Name().Resolve(ctx, aappns)
	if err != nil {
		return nil, fmt.Errorf("AAppns %v Resolve failed", aappns )
	}

	ap = &aapp{
		State:AAppStat_Unkown,
		CreateTime:time.Now().Unix(),
	}

	//Create MFS
	vroot, err := api.ResolveNode(ctx, path)
	if err != nil {
		return nil, err
	}

	pbnode, ok := vroot.(*dag.ProtoNode)
	if !ok {
		return nil, dag.ErrNotProtobuf
	}

	ap.VMFS, err = mfs.NewRoot(context.Background(), ind.DAG, pbnode, nil)
	if err != nil {
		return nil, err
	}

	if fsn, err := mfs.Lookup(ap.VMFS, "/Evn/info.json"); err != nil {
		return nil, errors.New(`"/Evn/info.json" not search file or directory`)
	} else {

		fi, ok := fsn.(*mfs.File)
		if !ok {
			return nil, errors.New(`"/Evn/info.json" was not a file`)
		}

		rfd, err := fi.Open(mfs.Flags{Read:true, Write:true})
		if err != nil {
			return nil, err
		}
		defer rfd.Close()

		filen, err := rfd.Size()
		if err != nil {
			return nil, err
		}

		r := io.LimitReader(rfd, filen)
		bs, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(bs, &ap.Info); err != nil {
			return nil, errors.New(`"/Evn/info.json" Unmarshal failed`)
		}
	}

	return ap, nil
}