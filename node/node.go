package node

import "ethereum-crawler/model"

type Node interface {
	Sync(start int64, x chan model.Account)
}
