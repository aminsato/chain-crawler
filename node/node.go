package node

import "chain-crawler/model"

type Node interface {
	Sync(start int64, x chan model.Account) error
	FirstBlock() int64
}
