package property

import (
	"errors"
	ADB "github.com/ayachain/go-aya-alvm-adb"
	"github.com/ipfs/go-mfs"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
)

type PropertyAPI interface {

	AvailBalanceOf( key []byte ) ( AvailRecord, error )
	FrozenBalanceOf( key []byte ) ( FrozenRecord, error)
	VotingRightOf( key []byte ) ( VotingRightRecord, error)

	AvailBalanceMove( from, to []byte, value *big.Int ) ( fromblc, toblc AvailRecord, err error )
	VotingRightMove( from, to []byte, value *big.Int) ( fromblc, toblc VotingRightRecord, err error )
}

var (
	addressNotFoundError = errors.New("target address not found")
	notEnoughError = errors.New("not enough balance")
	valueMustBePositiveNumberError = errors.New("value must be a positive number")
)

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

	if err := bn.Decode(bnc); err != nil {
		return bn, err
	}
	return bn, nil
}

func (api *aProperty) FrozenBalanceOf( key []byte ) ( r FrozenRecord, err error ) {

	vbs, err := api.frozenDB.Get( key, nil )
	if err != nil {
		return FrozenRecord{}, err
	}

	if err := r.Decode(vbs); err != nil {
		return r, err
	}

	return r, nil
}

func (api *aProperty) VotingRightOf( key []byte ) ( r VotingRightRecord, err error) {

	it := api.votingDB.NewIterator( &util.Range{Start:key}, nil )

	total := &big.Int{}

	for it.Next() {

		sbs := it.Value()
		sbn := &big.Int{}
		sbn.SetBytes(sbs)

		total.Add(total, sbn)
	}

	r.Set(total)
	return r, nil
}

func (api *aProperty) AvailBalanceMove( from, to []byte, value *big.Int ) ( fromblc, toblc AvailRecord, err error ) {

	if value.Cmp( new( big.Int ).SetInt64(0) ) <= 0 {
		return fromblc, toblc, valueMustBePositiveNumberError
	}

	fexist, err := api.availDB.Has(from, nil)
	if err != nil {
		return fromblc, toblc, err
	}

	if !fexist {
		return fromblc, toblc, addressNotFoundError
	}

	texist, err := api.availDB.Has(to, nil)
	if err != nil {
		return fromblc, toblc, err
	}

	fromblc, err = api.AvailBalanceOf(from)
	if err != nil {
		return fromblc, toblc, err
	}

	//fromVoteCount, err := api.VotingRightOf(from)
	if err != nil {
		return fromblc, toblc, err
	}

	if fromblc.Int.Cmp(value) < 0 {
		return fromblc, toblc, notEnoughError
	}

	if texist {
		toblc, err = api.AvailBalanceOf(to)
		if err != nil {
			return fromblc, toblc, err
		}
	}

	fromblc.Int.Sub( fromblc.Int, value)
	toblc.Int.Add(toblc.Int, value)

	batch := &leveldb.Batch{}
	batch.Put( from, fromblc.Bytes() )
	batch.Put( to, toblc.Bytes() )

	if err := api.availDB.Write(batch, nil); err != nil {
		return fromblc, toblc, err
	}

	return fromblc, toblc, err
}

func (api *aProperty) VotingRightMove( from, to []byte, value *big.Int) ( fromblc, toblc VotingRightRecord, err error ) {
	return VotingRightRecord{}, VotingRightRecord{}, nil
}