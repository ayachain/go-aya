package merger

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	VDBComm "github.com/ayachain/go-aya/vdb/common"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"strings"
)

type CVFSMerger interface {

	VDBComm.RawDBCoder

	GetBatchMap() map[string]*leveldb.Batch

	Put( dbKey string, k []byte, v []byte )

	Del( dbKey string, k []byte )

	Upload( ind *core.IpfsNode ) cid.Cid

}

type aCVFSMerger struct {

	CVFSMerger

	batchs map[string]*leveldb.Batch
}

func NewMerger() CVFSMerger {
	return &aCVFSMerger{
		batchs:make(map[string]*leveldb.Batch),
	}
}

func (tbg *aCVFSMerger) GetBatchMap() map[string]*leveldb.Batch{
	return tbg.batchs
}

// Byte 0 - 7 		: Header json bytes content len
// Byte 8 - HeadLen : json bytes content
// Bate .... 		: Batch dump bytes
func (tbg *aCVFSMerger) Encode() []byte {

	batchBuff := bytes.NewBuffer([]byte{})

	var head []string

	for _, k := range VDBComm.StorageDBPaths {

		if batch, exist := tbg.batchs[k]; exist {

			if batch != nil {

				batchbs := batch.Dump()

				head = append(head, fmt.Sprintf("%s:%d", k, len(batchbs)))

				batchBuff.Write(batchbs)

			}
		}
	}

	headBs, err := json.Marshal(head)
	if err != nil {
		return nil
	}

	logbuf := bytes.NewBuffer( VDBComm.BigEndianBytes( uint64(len(headBs))) )
	logbuf.Write( headBs )

	_, err = batchBuff.WriteTo(logbuf)
	if err != nil {
		return nil
	}

	return logbuf.Bytes()
}

func (tgb *aCVFSMerger) Decode( bs []byte ) error {

	buff := bufio.NewReader( bytes.NewReader(bs) )

	hlenbs := make([]byte, 8)
	if _, err := buff.Read(hlenbs); err != nil {
		return err
	}

	hlen := binary.BigEndian.Uint64(hlenbs)
	hcbs := make([]byte, hlen)
	if _, err := buff.Read(hcbs); err != nil {
		return err
	}

	var head []string
	if err := json.Unmarshal(hcbs, &head); err != nil {
		return err
	}

	for _, h := range head {

		arr := strings.SplitN(h,":", 2)
		k := arr[0]

		v, err := strconv.ParseUint( arr[1], 10,64)
		if err != nil {
			return err
		}

		tbs := make([]byte, v)
		if _, err := buff.Read(tbs); err != nil {
			return err
		}

		batch := &leveldb.Batch{}
		if err := batch.Load( tbs ); err != nil {
			return err
		}

		tgb.batchs[k] = batch
	}

	return nil
}

func (tgb *aCVFSMerger) Put( dbKey string, k []byte, v []byte ) {

	batch, exist := tgb.batchs[dbKey]
	if !exist {
		batch = &leveldb.Batch{}
		tgb.batchs[dbKey] = batch
	}

	batch.Put(k, v)
}

func (tgb *aCVFSMerger) Del( dbKey string, k []byte ) {

	batch, exist := tgb.batchs[dbKey]
	if !exist {
		batch := &leveldb.Batch{}
		tgb.batchs[dbKey] = batch
	}

	batch.Delete(k)
}

func (tgb *aCVFSMerger) Upload( ind *core.IpfsNode ) cid.Cid {

	dumpbs := tgb.Encode()

	dblk := blocks.NewBlock(dumpbs)

	_ = ind.Blocks.AddBlock(dblk)

	return dblk.Cid()
}