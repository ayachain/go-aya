package worker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	AvdbComm "github.com/ayachain/go-aya/vdb/common"
	"github.com/syndtr/goleveldb/leveldb"
)

// In order to ensure that the final task group can write concurrently, the
// target of each database should have as many as one task. The same task cannot
// be included in the group, and if so, it needs to be merged.
type TaskBatchGroup struct {
	AvdbComm.RawDBCoder
	batchs map[string]*leveldb.Batch
}

func NewGroup() *TaskBatchGroup {
	return &TaskBatchGroup{

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

	head := map[string]int{}

	for k, batch := range tbg.batchs {
		batchbs := batch.Dump()
		head[k] = len(batchbs)
		batchBuff.Write(batchbs)
	}

	headBs, err := json.Marshal(head)
	if err != nil {
		return nil
	}

	logbuf := bytes.NewBuffer( AvdbComm.BigEndianBytes( uint64(len(headBs))) )
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

	head := map[string]int{}
	if err := json.Unmarshal(hcbs, head); err != nil {
		return err
	}

	for k, v := range head {

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
		batch := &leveldb.Batch{}
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