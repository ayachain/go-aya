//
// When the watchdog passes down trusted messages, the notary coordinates his
// staff to process them, and eventually generates one or more "Batch" objects
// for the database, and finally writes one-time constraints by the next link,
// that is, either all writes, or all discards.
//
//
// We recommend implementing one or more workers to handle different things,
// and then distributing these messages through a simple factory model.
//

package worker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	ADog "github.com/ayachain/go-aya/consensus/core/watchdog"
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

func (tbg *TaskBatchGroup) GetBatchs() map[string]*leveldb.Batch{
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


type TaskBatchGroupBinder interface {

	Put ( dbkey string, k []byte, v []byte )

	Del ( dbkey string, k []byte, v []byte )

	GetTask( dbkey string ) *leveldb.Batch

	AppendTask( dbkey string, batch *leveldb.Batch )

	AppendGroup( group *TaskBatchGroup )

}


type TaskWorker interface {

	Processing( dogs *ADog.MsgFromDogs ) (*TaskBatchGroup, error)

}