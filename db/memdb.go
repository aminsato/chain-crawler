package db

import (
	"errors"
)

var memdbNotFoundErr = errors.New("key not found in the map")

type MemClient[C any] struct {
	db map[string]C
}

func NewMemDB[c any]() *MemClient[c] {
	db := make(map[string]c)
	cli := &MemClient[c]{
		db: db,
	}
	return cli
}

func (cli *MemClient[T]) Add(key string, val T) error {
	cli.db[key] = val
	return nil
}

func (cli *MemClient[T]) IsNotFoundError(err error) bool {
	return errors.Is(err, memdbNotFoundErr)
}

func (cli *MemClient[T]) Delete(key string) error {
	delete(cli.db, key)
	return nil
}

func (cli *MemClient[T]) Get(key string) (T, error) {
	value, found := cli.db[key]
	if !found {
		return value, memdbNotFoundErr
	}
	return value, nil
}

func (cli *MemClient[T]) Records(_ *string, _ *string, allRecords chan DBItem[T]) (err error) {
	defer close(allRecords)
	for k, v := range cli.db {
		allRecords <- DBItem[T]{
			Key:   k,
			Value: v,
		}
	}
	return nil
}
