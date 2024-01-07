package db

type DB[T any] interface {
	Add(key string, val T) error
	Delete(key string) error
	Get(key string) (T, error)
	Records(startKey *string, endKey *string) (allRecords map[string]T, err error)
	IsNotFoundError(err error) bool
}
