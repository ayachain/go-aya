package common

import (
	"encoding/binary"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
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

func OpenDB( dir *mfs.Directory ) (*leveldb.DB, storage.Storage, error) {

	dbstroage := ADB.NewMFSStorage(dir)
	if dbstroage == nil {
		panic("create adb storage expected")
	}

	db, err := leveldb.Open(dbstroage, &opt.Options{})

	if err != nil {
		return nil,nil,err
	}

	return db, dbstroage, nil
}

func OpenExistedDB( dir *mfs.Directory, path string, rdonly bool ) (*leveldb.DB, storage.Storage) {

	dbstroage := ADB.NewMFSStorage(dir)
	if dbstroage == nil {
		panic("create adb storage expected")
	}

	db, err := leveldb.Open(dbstroage, &opt.Options{})

	if err != nil {
		panic(err)
	}

	if rdonly {
		_ = db.SetReadOnly()
	}

	return db, dbstroage
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

func CacheHas( originDB *leveldb.DB, cacheDB *leveldb.DB, key []byte ) (bool, error) {

	exist, err := cacheDB.Has(key, nil)
	if err != nil {
		return false, err
	}

	if !exist {

		oexist, err := originDB.Has(key,nil)
		if err != nil {
			return false, err
		}

		return oexist, nil
	}

	return exist, nil
}

func CacheGet( originDB *leveldb.DB, cacheDB *leveldb.DB, key []byte ) ([]byte, error) {

	exist, err := cacheDB.Has(key, nil)
	if err != nil {
		return nil, err
	}

	if !exist {

		v, err := originDB.Get(key, nil)
		if err != nil {
			return nil, err
		}

		if err := cacheDB.Put(key, v, nil); err != nil {
			return nil, err
		}

		return v, nil

	} else {

		return cacheDB.Get(key, nil)

	}

}