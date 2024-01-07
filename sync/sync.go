package sync

import (
	"context"
	"ethereum-crawler/db"
	"ethereum-crawler/model"
	"ethereum-crawler/node"
	"ethereum-crawler/utils"
)

const LastHeightKey = "lastheight"

type Sync struct {
	node   node.Node
	db     db.DB[model.Account]
	height int64
	ctx    context.Context
	log    utils.SimpleLogger
}

func New(ctx context.Context, node node.Node, db db.DB[model.Account], log utils.SimpleLogger) (*Sync, error) {
	lastItem, err := db.Get(LastHeightKey)
	if err != nil {
		if db.IsNotFoundError(err) {
			lastItem = model.Account{
				Height: 90000,
			}
			err = db.Add(LastHeightKey, lastItem)
		}
	}
	if err != nil {
		return nil, err
	}

	return &Sync{
		node:   node,
		db:     db,
		ctx:    ctx,
		height: lastItem.Height,
		log:    log,
	}, nil
}

func (s *Sync) Start() error {
	result := make(chan model.Account, 10)
	s.node.Sync(s.height, result)
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case result := <-result:
			account, err := s.db.Get(result.Address)
			if err != nil && !s.db.IsNotFoundError(err) {
				return err
			}
			if account.Height < result.Height || (account.Height == result.Height && account.TxIndex < result.TxIndex) {
				account.Address = result.Address
				account.TotalPaidFee += result.TotalPaidFee
				account.Height = result.Height
				account.TxIndex = result.TxIndex
				err = s.db.Add(account.Address, account)
				if err != nil {
					return err
				}
			}
			if result.Height%10 == 0 {
				lastItem := model.Account{
					Height:  result.Height,
					TxIndex: result.TxIndex,
				}
				err = s.db.Add(LastHeightKey, lastItem)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
