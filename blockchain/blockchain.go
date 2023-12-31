package blockchain

import (
	"ethereum-crawler/db"
	"ethereum-crawler/utils"
)

type Blockchain struct {
	database db.DB
	log      utils.SimpleLogger
}

func New(database db.DB, log utils.SimpleLogger) *Blockchain {
	return &Blockchain{
		database: database,
		log:      log,
	}
}

//func (b *Blockchain) SaveSendersFee() {
//	for i := 0; ; i++ {
//		blockNumber := new(big.Int).SetInt64(int64(i))
//		block, err := b.node.GetBlockNumber(*blockNumber)
//		if err != nil {
//			println(err.Error())
//
//		}
//		for k := 0; k < block.Transactions().Len(); k++ {
//			transaction, err := b.node.GetTransaction(block.Hash(), uint(k))
//			if err != nil {
//				println(err.Error())
//			}
//			sender, err := b.node.GetTransactionSender(transaction, block.Hash(), uint(k))
//			if err != nil {
//				println(err.Error())
//			}
//			receipt, err := b.node.GetTransactionReceipt(transaction.Hash())
//			if err != nil {
//				println(err.Error())
//			}
//
//			b.database.SaveSederBalance(sender.String(), transaction.GasPrice().Uint64()*receipt.GasUsed)
//		}
//
//	}
//}
