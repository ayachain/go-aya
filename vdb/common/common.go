package common

import (
	"encoding/binary"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
)

type RawDBCoder interface {
	Encode() []byte
	Decode([]byte) error
}


type AMessageEncode interface {
	RawMessageEncode() []byte
	RawMessageDecode( bs []byte ) error
}

type RawSigner interface {
	RawSignEncode( account accounts.Account ) ([]byte, error)
	RawVerifyDecode( bs []byte ) error
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

func BigEndianBytesUint32 (n uint32 ) []byte {
	enc := make([]byte, 4)
	binary.BigEndian.PutUint32(enc, n)
	return enc
}

func BigEndianBytesUint16 ( n uint16 ) []byte {
	enc := make([]byte, 2)
	binary.BigEndian.PutUint16(enc, n)
	return enc
}

func OpenExistedDB( dir *mfs.Directory, path string ) *leveldb.DB {

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

func LookupDBPath( root *mfs.Root, path string ) (*mfs.Directory, error) {

	nd, err := mfs.Lookup(root, path)

	if err != nil {

		err := mfs.Mkdir(root, path, mfs.MkdirOpts{ Mkparents:true, Flush:false })
		if err != nil {
			return nil, err
		}

		nd, err = mfs.Lookup(root, path)
		if err != nil {
			return nil, err
		}

	}

	dir, ok := nd.(*mfs.Directory)
	if !ok {
		return nil, mfs.ErrInvalidChild
	}

	return dir, nil
}