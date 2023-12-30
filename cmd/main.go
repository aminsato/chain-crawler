package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/pebble"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
)

var infuraURL = "https://mainnet.infura.io/v3/8ae89b94ba6640cb8f9d1c42b53f21ee"

func main() {
	// Connect to a local Ethereum node (make sure it's running)
	client, err := rpc.Dial(infuraURL)
	if err != nil {
		log.Fatal(err)
	}

	// Open or create a Pebble database
	db, err := pebble.Open("ethereum_db", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Replace the following parameters with your filter criteria
	fromBlock := big.NewInt(0)
	toBlock := big.NewInt(1000000)
	addresses := []common.Address{common.HexToAddress("0x1234567890123456789012345678901234567890")}

	// Create filter options
	filterOpts := ethereum.FilterQuery{
		Addresses: addresses,
		FromBlock: fromBlock,
		ToBlock:   toBlock,
	}

	// Subscribe to logs
	logsCh := make(chan types.Log)
	ctx := context.Background()
	sub, err := client.EthSubscribe(ctx, logsCh, "logs", filterOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	fmt.Println("Listening for logs...")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case log := <-logsCh:
			if err := saveLogToDB(db, log); err != nil {
				fmt.Println("Error saving log to DB:", err)
			}
			printLog(log)
		}
	}
}

func saveLogToDB(db *pebble.DB, log types.Log) error {
	// Convert log to JSON
	logJSON, err := json.Marshal(log)
	if err != nil {
		return err
	}

	// Use block hash as key
	key := []byte(log.BlockHash.Hex())

	// Save log data to the Pebble database
	return db.Set(key, logJSON, pebble.Sync)
}

func printLog(log types.Log) {
	// You can customize the output or processing of the log data here
	fmt.Printf("BlockHash: %s\n", log.BlockHash.Hex())
	fmt.Printf("BlockNumber: %s\n", log.BlockNumber)
	fmt.Printf("TxHash: %s\n", log.TxHash.Hex())
	fmt.Printf("TxIndex: %d\n", log.TxIndex)
	fmt.Printf("Index: %d\n", log.Index)
	fmt.Printf("Address: %s\n", log.Address.Hex())
	fmt.Printf("Data: %s\n", log.Data)
	fmt.Println("--------------------------------------")
}
