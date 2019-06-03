package property

import (
	"encoding/json"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
)

type PropertyAPI interface {

	AvailBalanceOf( key []byte ) ( AvailRecord, error )

	FrozenBalanceOf( key []byte ) ( FrozenRecord, error)

	VotingRightOf( key []byte ) ( VotingRightRecord, error)

	AvailBalanceMove( from, to []byte, value *big.Int ) ( fromblc, toblc AvailRecord, err error )

	BallotCountMove( from, to []byte, value *big.Int) ( fromblc, toblc VotingRightRecord, err error )

}

const (
	availDBPath = "/db/avail"
	frozenDBPath = "/db/frozen"
	votingDBPath = "/db/ballot"
)

type aProperty struct {

	PropertyAPI
	*mfs.Root

	availDB *leveldb.DB
	frozenDB *leveldb.DB
	votingDB *leveldb.DB
}

func checkPathAndCreatePathAndDB( root *mfs.Root, path string ) *leveldb.DB {

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

func CreatePropertyAPI( rootref *mfs.Root ) PropertyAPI {

	api := &aProperty{
		Root:rootref,
	}

	api.availDB = checkPathAndCreatePathAndDB(rootref, availDBPath)
	api.frozenDB = checkPathAndCreatePathAndDB(rootref, frozenDBPath)
	api.votingDB = checkPathAndCreatePathAndDB(rootref, votingDBPath)

	return api
}


func (api *aProperty) AvailBalanceOf( key []byte ) ( AvailRecord, error ) {

	bnc, err := api.availDB.Get(key, nil)
	if err != nil {
		return AvailRecord{}, err
	}

	bn := AvailRecord{}
	bn.SetBytes(bnc)
	return bn, nil

}


func (api *aProperty) FrozenBalanceOf( key []byte ) ( r FrozenRecord, err error ) {

	vbs, err := api.frozenDB.Get( key, nil )
	if err != nil {
		return FrozenRecord{}, err
	}

	if err := json.Unmarshal(vbs, &r); err != nil {
		return FrozenRecord{}, err
	}

	return r, nil
}


func (api *aProperty) VotingRightOf( key []byte ) (*big.Int, error) {

	

}