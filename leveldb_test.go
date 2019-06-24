package main

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"testing"
)

func TestOpenAndClose( t *testing.T ) {

	stor, err := storage.OpenFile("/Users/apple/Desktop/SSSS/Test", false)

	db, err := leveldb.Open(stor, nil)
	if err != nil {
		t.Error(err)
	}

	if err := stor.Close(); err != nil {
		t.Error(err)
	}

	if err := db.Put([]byte("Key"), []byte("Value"),nil); err != nil {
		t.Error(err)
	}

	if err := db.Close(); err != nil {
		t.Error(err)
	}

	db, err = leveldb.OpenFile("/Users/apple/Desktop/SSSS/Test", nil)
	if err != nil {
		t.Error(err)
	}

	if _, err := db.Get([]byte("Key"), nil); err != nil {
		t.Error(err)
	}

	if err := db.Close(); err != nil {
		t.Error(err)
	}

}