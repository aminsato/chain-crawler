package node

import (
	"context"
	"ethereum-crawler/model"
	"ethereum-crawler/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"time"
)

type EthNode struct {
	ctx     context.Context
	client  *ethclient.Client
	log     utils.SimpleLogger
	blockCh chan *types.Block
	height  int64
}

func NewEthNode(ctx context.Context, nodeAddress string, log *utils.ZapLogger) (*EthNode, error) {
	client, err := ethclient.DialContext(ctx, nodeAddress)
	if err != nil {
		return nil, err
	}

	return &EthNode{
		ctx:     ctx,
		client:  client,
		log:     log,
		blockCh: make(chan *types.Block, 4000),
	}, nil
}
func (e *EthNode) getBlockNumber(block_number big.Int) (*types.Block, error) {
	block, err := e.client.BlockByNumber(e.ctx, &block_number)
	if err != nil {
		e.log.Errorw("Error to get block by number:", err)
	}
	return block, err

}
func (e *EthNode) getTransaction(blockHash common.Hash, index uint) (*types.Transaction, error) {
	transaction, err := e.client.TransactionInBlock(e.ctx, blockHash, index)

	if err != nil {
		e.log.Errorw("Error to get transaction in block:", err)
	}
	return transaction, err
}
func (e *EthNode) getTransactionSender(transaction *types.Transaction, blockHash common.Hash, index uint) (common.Address, error) {
	sender, err := e.client.TransactionSender(e.ctx, transaction, blockHash, index)
	if err != nil {
		e.log.Errorw("error to get transaction sender:", err)
	}
	return sender, err
}
func (e *EthNode) getTransactionReceipt(transactionHash common.Hash) (*types.Receipt, error) {
	receipt, err := e.client.TransactionReceipt(e.ctx, transactionHash)
	if err != nil {
		e.log.Errorw("error to get transaction receipt:", err)
	}
	return receipt, err
}

func (e *EthNode) getSenderBakance(sender common.Address) (*big.Int, error) {
	balance, err := e.client.BalanceAt(e.ctx, sender, nil)
	if err != nil {
		e.log.Errorw("error to get sender balance:", err)
	}
	return balance, err
}

func (e *EthNode) Sync(start int64, result chan model.Account) {
	go e.fetchBlocks(start)
	go e.fetchTransactions(result)
}

func (e *EthNode) fetchBlocks(start int64) error {
	for i := start; ; i++ {
		//check context and it its done or there is an error, skip the loop
		blockNumber := new(big.Int).SetInt64(i)
		block, err := e.getBlockNumber(*blockNumber)
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
		//TODO: when synced, we should wait for new blocks
	}
	return nil
}

func (e *EthNode) fetchTransactions(result chan model.Account) error {
	for {
		select {
		case <-e.ctx.Done():
			return nil
		case block := <-e.blockCh:
			for k := 0; k < block.Transactions().Len(); k++ {
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
				result <- model.Account{
					Address:      sender.String(),
					TotalPaidFee: transaction.GasPrice().Uint64() * receipt.GasUsed,
					Height:       block.Number().Int64(),
					TxIndex:      k,
				}
			}
		}
	}
	return nil
}
