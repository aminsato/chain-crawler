package sync

import (
	"context"

	"chain-crawler/db"
	"chain-crawler/model"
	"chain-crawler/node"
	"chain-crawler/utils"
)

type Sync struct {
	node   node.Node
	db     db.DB[model.Account]
	height int64
	ctx    context.Context
	log    utils.SimpleLogger
	total  model.Account
}

func New(ctx context.Context, node node.Node, database db.DB[model.Account], log utils.SimpleLogger) (*Sync, error) {
	lastItem, err := database.Get(db.LastHeightKey)
	if err != nil {
		if database.IsNotFoundError(err) {
			lastItem = model.Account{
				LastHeight: node.FirstBlock(),
			}
			err = database.Add(db.LastHeightKey, lastItem)
		}
	}
	if err != nil {
		return nil, err
	}

	return &Sync{
		node:   node,
		db:     database,
		ctx:    ctx,
		height: lastItem.LastHeight,
		log:    log,
		total:  lastItem,
	}, nil
}

func (s *Sync) Start() error {
	result := make(chan model.Account, 10)
	errCh := make(chan error)
	go func() {
		err := s.node.Sync(s.height, result)
		if err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case err := <-errCh:
			return err
		case result := <-result:
			account, err := s.db.Get(result.Address)
			if err != nil && !s.db.IsNotFoundError(err) {
				return err
			}
			if account.LastHeight < result.LastHeight || (account.LastHeight == result.LastHeight && account.TxIndex < result.TxIndex) {
				account.Address = result.Address
				account.TotalPaidFee += result.TotalPaidFee
				account.LastHeight = result.LastHeight
				account.TxIndex = result.TxIndex
				account.IsContract = result.IsContract
				if account.FirstHeight == 0 {
					account.FirstHeight = result.LastHeight
				}
				err = s.db.Add(account.Address, account)
				if err != nil {
					return err
				}
			}
			s.total.TotalPaidFee += result.TotalPaidFee
			if result.LastHeight%10 == 0 {
				lastItem := model.Account{
					LastHeight:   result.LastHeight,
					TxIndex:      result.TxIndex,
					FirstHeight:  result.LastHeight,
					TotalPaidFee: s.total.TotalPaidFee,
					IsContract:   result.IsContract,
				}
				err = s.db.Add(db.LastHeightKey, lastItem)
				if err != nil {
					return err
				}
			}
		}
	}
}
