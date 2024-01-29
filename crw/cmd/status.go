/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"chain-crawler/db"
	"chain-crawler/model"
	"chain-crawler/utils"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get last height of the chain",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		log, err := utils.NewZapLogger(cfg.LogLevel, cfg.Colour)
		if err != nil {
			log.Fatalf(err.Error())
		}
		dbPath := cfg.DatabasePath
		switch cfg.Chain {
		case "eth":
			dbPath = dbPath + "/ethereum_db"
		case "bsc":
			dbPath = dbPath + "/binance_db"
		default:
			log.Errorw("chain not supported")
			return
		}
		chainDb, err := db.NewLevelDB[model.Account](dbPath)
		if err != nil {
			log.Error("Error opening chainDb", err)
			log.Fatalf(err.Error())
		}
		res, err := chainDb.Get(db.LastHeightKey)
		if err != nil && !chainDb.IsNotFoundError(err) {
			log.Errorw(err.Error())
		}
		log.Infow(fmt.Sprintf("%v", res.LastHeight))
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.Parent().PreRun(cmd.Parent(), args)
	},
}

func init() {
	statusCmd.Flags().StringVarP(&cfg.Chain, chainF, "c", defaultChain, chainUsage)
	rootCmd.AddCommand(statusCmd)
}
