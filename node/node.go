package node

import (
	"context"
	"ethereum-crawler/blockchain"
	"ethereum-crawler/config"
	"ethereum-crawler/db"
	"ethereum-crawler/utils"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

//	type Node interface {
//		Subscribe(from, to int64) error
//		GetTransaction(blockHash common.Hash, index uint) (*types.Transaction, error)
//		GetBlockNumber(block_number big.Int) (*types.Block, error)
//		GetTransactionSender(transaction *types.Transaction, blockHash common.Hash, index uint) (common.Address, error)
//		GetTransactionReceipt(hash common.Hash) (*types.Receipt, error)
//	}
type EthNode struct {
	ctx        context.Context
	client     *ethclient.Client
	database   db.DB
	blockchain *blockchain.Blockchain
	logsCh     chan types.Log
	blockCh    chan *types.Block
}

func New(cfg *config.Config, ctx context.Context) (*EthNode, error) {
	client, err := ethclient.DialContext(ctx, cfg.NodeAddress)
	if err != nil {
		return nil, err
	}

	log, err := utils.NewZapLogger(cfg.LogLevel, cfg.Colour)

	//dbLog, err := utils.NewZapLogger(utils.ERROR, cfg.Colour)

	database, err := db.New(cfg.DatabasePath, log)
	if err != nil {
		fmt.Errorf("open DB: %w", err)
	}

	//chain := blockchain.New(*database, dbLog)
	return &EthNode{
		ctx:      ctx,
		client:   client,
		database: *database,
		logsCh:   make(chan types.Log, 10000),
		blockCh:  make(chan *types.Block, 4000),
	}, nil
}

func (e *EthNode) Subscribe(from, to int64) error {
	fromBlock := big.NewInt(from)
	toBlock := big.NewInt(to)
	filterOpts := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
	}
	sub, err := e.client.SubscribeFilterLogs(e.ctx, filterOpts, e.logsCh)
	if err != nil {
		return err
	}
	_ = sub
	for {
		select {
		case <-e.ctx.Done():
			fmt.Println("Done")
			return nil
		case log := <-e.logsCh:
			tx, err := e.GetTransaction(log.BlockHash, log.TxIndex)
			if err != nil {
				return err
			}
			sender, err := e.GetTransactionSender(tx, log.BlockHash, log.TxIndex)
			if err != nil {
				println(err.Error())
				return err
			}
			receipt, err := e.GetTransactionReceipt(tx.Hash())
			if err != nil {
				println(err.Error())
				return err
			}
			err = e.database.SaveSederBalance(sender.String(), tx.GasPrice().Uint64()*receipt.GasUsed)
			if err != nil {
				println(err.Error())
				return err
			}
		}
	}
}

func (e *EthNode) GetBlockNumber(block_number big.Int) (*types.Block, error) {

	//e.client.SubscribeFilterLogs()
	block, err := e.client.BlockByNumber(e.ctx, &block_number)
	if err != nil {
		println("Error to get block by number:", err)
	}
	return block, err

}
func (e *EthNode) GetTransaction(blockHash common.Hash, index uint) (*types.Transaction, error) {
	transaction, err := e.client.TransactionInBlock(e.ctx, blockHash, index)

	if err != nil {
		println("Error to get transaction in block:", err)
	}
	return transaction, err
}
func (e *EthNode) GetTransactionSender(transaction *types.Transaction, blockHash common.Hash, index uint) (common.Address, error) {
	sender, err := e.client.TransactionSender(e.ctx, transaction, blockHash, index)
	if err != nil {
		fmt.Println("error to get transaction sender:", err)
	}
	return sender, err
}
func (e *EthNode) GetTransactionReceipt(transactionHash common.Hash) (*types.Receipt, error) {
	receipt, err := e.client.TransactionReceipt(e.ctx, transactionHash)
	if err != nil {
		fmt.Println("error to get transaction receipt:", err)
	}
	return receipt, err
}

func (e *EthNode) getSenderBakance(sender common.Address) (*big.Int, error) {
	balance, err := e.client.BalanceAt(e.ctx, sender, nil)
	if err != nil {
		fmt.Println("error to get sender balance:", err)
	}
	return balance, err
}

func (e *EthNode) StartFetch() {
	e.SaveSendersFee()
}

func (e *EthNode) SaveSendersFee() {
	for i := 1; ; i++ {
		blockNumber := new(big.Int).SetInt64(int64(i))
		block, err := e.GetBlockNumber(*blockNumber)
		if err != nil {
			println(err.Error())

		}
		for k := 0; k < block.Transactions().Len(); k++ {
			transaction, err := e.GetTransaction(block.Hash(), uint(k))
			if err != nil {
				println(err.Error())
			}
			sender, err := e.GetTransactionSender(transaction, block.Hash(), uint(k))
			if err != nil {
				println(err.Error())
			}
			receipt, err := e.GetTransactionReceipt(transaction.Hash())
			if err != nil {
				println(err.Error())
			}

			e.database.SaveSederBalance(sender.String(), transaction.GasPrice().Uint64()*receipt.GasUsed)
		}
	}
}

func (e *EthNode) FetchBlocks(start int64) error {
	for i := start; ; i++ {
		blockNumber := new(big.Int).SetInt64(int64(i))
		block, err := e.GetBlockNumber(*blockNumber)
		if err != nil {
			println(err.Error())
			return err
		}
		e.blockCh <- block
		//TODO: when synced, we should wait for new blocks
	}
	return nil
}

func (e *EthNode) FetchTransactions() error {
	for {
		select {
		case <-e.ctx.Done():
			fmt.Println("Done")
			return nil
		case block := <-e.blockCh:
			fmt.Printf("block number: %d, tx count:%d\n", block.Number(), block.Transactions().Len())
			for k := 0; k < block.Transactions().Len(); k++ {
				transaction, err := e.GetTransaction(block.Hash(), uint(k))
				if err != nil {
					println(err.Error())
				}
				sender, err := e.GetTransactionSender(transaction, block.Hash(), uint(k))
				if err != nil {
					println(err.Error())
				}
				receipt, err := e.GetTransactionReceipt(transaction.Hash())
				if err != nil {
					println(err.Error())
				}
				_ = receipt
				_ = sender
				/*	err = e.database.SaveSederBalance(sender.String(), transaction.GasPrice().Uint64()*receipt.GasUsed)
					if err != nil {
						println(err.Error())
					}*/
			}
		}
	}
	return nil
}
