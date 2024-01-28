package node

import (
	"context"
	"math/big"
	"time"

	"chain-crawler/model"
	"chain-crawler/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Binance Smart Chain
type BscNode struct {
	ctx           context.Context
	client        *ethclient.Client
	log           utils.SimpleLogger
	blockCh       chan *types.Block
	limiterForReq <-chan time.Time
}

func NewBscNode(ctx context.Context, nodeAddress string, channelSize int, rps int, log *utils.ZapLogger) (Node, error) {
	client, err := ethclient.DialContext(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	return &BscNode{
		ctx:           ctx,
		client:        client,
		log:           log,
		blockCh:       make(chan *types.Block, channelSize),
		limiterForReq: time.NewTicker(time.Second / time.Duration(rps)).C,
	}, nil
}

func (e *BscNode) FirstBlock() int64 {
	return 1
}

func (e *BscNode) getBlockNumber(block_number big.Int) (*types.Block, error) {
	<-e.limiterForReq
	block, err := e.client.BlockByNumber(e.ctx, &block_number)
	if err != nil {
		e.log.Errorw("Error to get block by number:", err)
	}
	return block, err
}

func (e *BscNode) getTransaction(blockHash common.Hash, index uint) (*types.Transaction, error) {
	<-e.limiterForReq
	transaction, err := e.client.TransactionInBlock(e.ctx, blockHash, index)
	if err != nil {
		e.log.Errorw("Error to get transaction in block:", err)
	}
	return transaction, err
}

func (e *BscNode) getTransactionSender(transaction *types.Transaction, blockHash common.Hash, index uint) (common.Address, error) {
	<-e.limiterForReq
	sender, err := e.client.TransactionSender(e.ctx, transaction, blockHash, index)
	if err != nil {
		e.log.Errorw("error to get transaction sender:", err)
	}
	return sender, err
}

func (e *BscNode) getTransactionReceipt(transactionHash common.Hash) (*types.Receipt, error) {
	<-e.limiterForReq
	receipt, err := e.client.TransactionReceipt(e.ctx, transactionHash)
	if err != nil {
		e.log.Errorw("error to get transaction receipt:", err)
	}
	return receipt, err
}

func (e *BscNode) Sync(start int64, result chan model.Account) error {
	errCh := make(chan error)
	go func() {
		err := e.fetchBlocks(start)
		if err != nil {
			e.log.Errorw(err.Error())
			errCh <- err
		}
	}()
	go func() {
		err := e.fetchTransactions(result)
		if err != nil {
			e.log.Errorw(err.Error())
			errCh <- err
		}
	}()
	select {
	case <-e.ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (e *BscNode) fetchBlocks(start int64) error {
	for i := start; ; i++ {
		// check context and it its done or there is an error, skip the loop
		blockNumber := new(big.Int).SetInt64(i)

		block, err := e.getBlockNumber(*blockNumber)
		e.log.Infow("fetchBlocks", "block number", block.Number().Int64())
		if err != nil {
			if err.Error() == "not found" {
				e.log.Debugw("Block not found", "block", i)
				i--
				time.Sleep(2 * time.Second)
				continue
			}
			e.log.Errorw(err.Error())
			return err
		}
		e.blockCh <- block
		// TODO: when synced, we should wait for new blocks
	}
}

func (e *BscNode) fetchTransactions(result chan model.Account) error {
	for {
		select {
		case <-e.ctx.Done():
			return nil
		case block := <-e.blockCh:
			for k := 0; k < block.Transactions().Len(); k++ {
				e.log.Infow("fetchTransactions", "block number", block.Number().Int64(), "tx index", k)
				transaction, err := e.getTransaction(block.Hash(), uint(k))
				if err != nil {
					e.log.Errorw(err.Error())
					return err
				}

				sender, err := e.getTransactionSender(transaction, block.Hash(), uint(k))
				if err != nil {
					e.log.Errorw(err.Error())
					return err
				}
				receipt, err := e.getTransactionReceipt(transaction.Hash())
				if err != nil {
					e.log.Errorw(err.Error())
					return err
				}

				contractAddress := func() string {
					if transaction.To() == nil {
						return receipt.ContractAddress.String()
					} else {
						return transaction.To().String()
					}
				}()

				isContractInteraction := len(transaction.Data()) > 0
				if isContractInteraction {
					result <- model.Account{
						Address:      contractAddress,
						TotalPaidFee: transaction.GasPrice().Uint64() * receipt.GasUsed,
						LastHeight:   block.Number().Int64(),
						TxIndex:      k,
						IsContract:   true,
					}
				}

				result <- model.Account{
					Address:      sender.String(),
					TotalPaidFee: transaction.GasPrice().Uint64() * receipt.GasUsed,
					LastHeight:   block.Number().Int64(),
					TxIndex:      k,
				}
			}
		}
	}
}
