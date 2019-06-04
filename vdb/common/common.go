package common

import (
	"encoding/binary"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
)

type RawDBCoder interface {
	Encode() []byte
	Decode([]byte) error
}

func LittleEndianBytes (number uint64) []byte {
	enc := make([]byte, 8)
	binary.LittleEndian.PutUint64(enc, number)
	return enc
}

func BigEndianBytes (number uint64) []byte {

	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc

}

func OpenStandardDB( root *mfs.Root, path string ) *leveldb.DB {

	nd, err := mfs.Lookup(root, path)

	if err != nil {

		err := mfs.Mkdir(root, path, mfs.MkdirOpts{ Mkparents:true, Flush:false })
		if err != nil {
			panic(err)
		}

		nd, err = mfs.Lookup(root, path)
		if err != nil {
			panic("can't created db")
		}

	}

	dir, ok := nd.(*mfs.Directory)
	if !ok {
		panic("path is exist but not a standard db dir")
	}

	dbstroage := ADB.NewMFSStorage(dir)
	if dbstroage == nil {
		panic("create adb storage expected")
	}

	db, err := leveldb.Open(dbstroage, nil)
	if err != nil {
		panic(err)
	}

	return db
}