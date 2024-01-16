package db

const LastHeightKey = "lastHeight"

type DBItem[T any] struct {
	Key   string
	Value T
}
type DB[T any] interface {
	Add(key string, val T) error
	Delete(key string) error
	Get(key string) (T, error)
	Records(startKey *string, endKey *string, allRecords chan DBItem[T]) (err error)
	IsNotFoundError(err error) bool
}
