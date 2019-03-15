package utils

import (
	"encoding/json"
	"github.com/ipfs/go-ipfs-api"
)

func GetHash(v interface{}) (hstr string, err error) {

	switch v.(type) {
	case string:
		hstr, err = shell.NewLocalShell().BlockPut([]byte(v.(string)),"", "sha2-256",32)

	case []byte:
		hstr, err = shell.NewLocalShell().BlockPut([]byte(v.([]byte)),"", "sha2-256",32)

	default:

		if bs, err := json.Marshal(v); err == nil {
			hstr, err = shell.NewLocalShell().BlockPut(bs, "", "sha2-256", 32)
		}

	}

	if err != nil {
		return "", err
	} else {
		return hstr,nil
	}

}

func GetHashBytes(v interface{}) (hbs[] byte, err error) {

	hstr, err := GetHash(v)

	if err != nil {
		return nil, err
	}

	return []byte(hstr), nil

}