package db

import (
	"encoding/binary"
	"github.com/cockroachdb/pebble"
	"sync"
)

const (
	// minCache is the minimum amount of memory in megabytes to allocate to pebble read and write caching.
	minCache = 8

	megabyte = 1 << 20
)

//sender address and fee

type DB struct {
	pebble *pebble.DB
	wMutex *sync.Mutex
}

func New(path string, logger pebble.Logger) (*DB, error) {
	pDB, err := newPebble(path, &pebble.Options{
		Logger: logger,
	})

	if err != nil {
		return nil, err
	}
	return pDB, nil
}
func newPebble(path string, options *pebble.Options) (*DB, error) {
	pDB, err := pebble.Open(path, options)
	if err != nil {
		return nil, err
	}
	return &DB{pebble: pDB, wMutex: new(sync.Mutex)}, nil
}

func (db *DB) SaveSederBalance(senderAddress string, sum uint64) error {
	senderBytes := []byte(senderAddress)

	//define uint64
	var balance uint64
	// Convert uint64 to byte slice for the sum
	value, _, err := db.pebble.Get(senderBytes)

	//*uint64 to uint64

	if err == nil {
		balance += binary.LittleEndian.Uint64(value)
	}

	return db.pebble.Set(senderBytes, value, pebble.Sync)
}
