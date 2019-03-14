package utils

import (
	"encoding/json"
	"github.com/ipfs/go-ipfs-api"
)

func GetHash(v interface{}) (hstr string, err error) {

	if bs, err := json.Marshal(v); err == nil {

		hstr, err := shell.NewLocalShell().BlockPut(bs,"sha2-256", "",-1)

		if err != nil {
			return "", err
		} else {
			return hstr,nil
		}

	} else {
		return "", err
	}
}

func GetHashBytes(v interface{}) (hbs[] byte, err error) {

	hstr, err := GetHash(v)

	if err != nil {
		return nil, err
	}

	return []byte(hstr), nil

}