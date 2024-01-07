package db

import (
	"encoding/json"
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type LvlClient[C any] struct {
	db *leveldb.DB
}

func NewLevelDB[c any](address string) (*LvlClient[c], error) {
	db, err := leveldb.OpenFile(address, nil)
	cli := &LvlClient[c]{
		db: db,
	}
	return cli, err
}

func (cli *LvlClient[T]) Add(key string, val T) error {
	bt, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = cli.db.Put([]byte(key), bt, nil)
	if err != nil {
		return err
	}
	return err
}
func (cli *LvlClient[T]) IsNotFoundError(err error) bool {
	return errors.Is(err, leveldb.ErrNotFound)
}

func (cli *LvlClient[T]) Delete(key string) error {
	err := cli.db.Delete([]byte(key), nil)
	if err != nil {
		return err
	}
	return err
}

func (cli *LvlClient[T]) Get(key string) (T, error) {
	var data T
	result, err := cli.db.Get([]byte(key), nil)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(result, &data)
	return data, err
}

func (cli *LvlClient[T]) Records(startKey *string, endKey *string) (allRecords map[string]T, err error) {
	allRecords = make(map[string]T)
	var rng *util.Range
	if startKey != nil && endKey != nil {
		rng = &util.Range{Start: []byte(*startKey), Limit: []byte(*endKey)}
	} else if startKey != nil {
		rng = &util.Range{Start: []byte(*startKey)}
	} else if endKey != nil {
		rng = &util.Range{Limit: []byte(*endKey)}
	}
	iter := cli.db.NewIterator(rng, nil)
	defer iter.Release()
	for iter.Next() {
		var value T
		key := iter.Key()
		val := iter.Value()
		if iter.Value() != nil {
			err := json.Unmarshal(val, &value)
			if err != nil {
				return nil, err
			}
		}
		allRecords[string(key)] = value
	}
	iter.Release()
	return allRecords, err
}

func (cli LvlClient[C]) Close() error {
	return cli.db.Close()
}
