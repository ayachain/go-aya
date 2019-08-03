package worker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	AVdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"strings"
)

// In order to ensure that the final task group can write concurrently, the
// target of each database should have as many as one task. The same task cannot
// be included in the group, and if so, it needs to be merged.
type TaskBatchGroup struct {
	AVdbComm.RawDBCoder
	batchs map[string]*leveldb.Batch
}

func NewGroup() *TaskBatchGroup {
	return &TaskBatchGroup{
		batchs:make(map[string]*leveldb.Batch),
	}
}

func (tbg *TaskBatchGroup) GetBatchMap() map[string]*leveldb.Batch{
	return tbg.batchs
}

// Byte 0 - 7 : Header json bytes content len
// Byte 8 - HeadLen : json bytes content
// Bate .... : Batch dump bytes
func (tbg *TaskBatchGroup) Encode() []byte {

	batchBuff := bytes.NewBuffer([]byte{})

	var head []string

	for _, k := range AVdbComm.StorageDBPaths {

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

	logbuf := bytes.NewBuffer( AVdbComm.BigEndianBytes( uint64(len(headBs))) )
	logbuf.Write( headBs )

	_, err = batchBuff.WriteTo(logbuf)
	if err != nil {
		return nil
	}

	return logbuf.Bytes()
}

func (tgb *TaskBatchGroup) Decode( bs []byte ) error {

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

	head := []string{}
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

func (tgb *TaskBatchGroup) Put( dbkey string, k []byte, v []byte ) {

	batch, exist := tgb.batchs[dbkey]
	if !exist {
		batch = &leveldb.Batch{}
		tgb.batchs[dbkey] = batch
	}

	batch.Put(k, v)
}

func (tgb *TaskBatchGroup) Del( dbkey string, k []byte ) {

	batch, exist := tgb.batchs[dbkey]
	if !exist {
		batch := &leveldb.Batch{}
		tgb.batchs[dbkey] = batch
	}

	batch.Delete(k)
}