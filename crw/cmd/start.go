/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"chain-crawler/config"

	grpc "chain-crawler/service/grpc/server"
	"chain-crawler/service/http"

	"chain-crawler/db"
	"chain-crawler/model"
	"chain-crawler/node"
	"chain-crawler/sync"
	"chain-crawler/utils"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	defaultNodeChanSize            = "10"
	defaultRequestPerSecond        = "10"
	defaultEthNodeAddress          = "your eth node address"
	defaultBscAddress              = "your bsc node address"
	defaultEthHttpPort      uint16 = 1080
	defaultEthGrpcPort      uint16 = 1082
	defaultBscHttpPort      uint16 = 1081
	defaultBscGrpcPort      uint16 = 1083
	defaultGRPC                    = false
	defaultChain                   = "eth,bsc"

	dbPathF           = "db-path"
	requestPerSecondF = "rps"
	ethNodeAddressF   = "eth-node-address"
	bscNodeAddressF   = "bsc-node-address"
	ethHTTPPortF      = "eth-service-port"
	ethGrpcPortF      = "eth-grpc-port"
	bscHTTPPortF      = "bsc-service-port"
	bscGrpcPortF      = "bsc-grpc-port"
	grpcF             = "grpc"
	chainF            = "chain"
	nodeChanSizeF     = "node-chan-size"
	logLevelF         = "log-level"

	dbPathUsage           = "Location of the database files."
	requestPerSecondUsage = "Maximum number of requests per second for gateway endpoints"
	nodeAddressUsage      = "The address of the node to connect to."
	ethHTTPPortUsage      = "The httpPort on which the HTTP server will listen for eth requests."
	bscHTTPPortUsage      = "The httpPort on which the HTTP server will listen for bsc requests."
	ethGrpcPortUsage      = "The grpcPort on which the grpc server will listen for eth requests."
	bscGrpcPortUsage      = "The grpcPort on which the grpc server will listen for bsc requests."
	chainUsage            = "The chains to crawl, use , to separate chains"
	nodeChanSizeUsage     = "The size of the channel that will be used to communicate with the eth node."
	logLevelFlagUsage     = "Options: debug, info, warn, error."
	grpcUsage             = "Enable grpc server"

	Version = "0.0.1"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [flags]]",
	Short: "start crawler",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.Parent().PreRun(cmd.Parent(), args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		log, err := utils.NewZapLogger(cfg.LogLevel, cfg.Colour)
		if err != nil {
			log.Fatalf("Error creating logger: %v", err)
		}
		if cfg.Chain == "" {
			log.Fatalf("chain flag is required")
		}
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(context.Background())
		g, ctx := errgroup.WithContext(ctx)

		go func() {
			<-quit
			cancel()
		}()
		var dbPath string
		var chainNode node.Node
		var httpPort uint16
		var grpcPort uint16
		for _, chain := range strings.Split(cfg.Chain, ",") {
			switch chain {
			case "eth":
				dbPath = fmt.Sprintf("%s/%s", cfg.DatabasePath, "/ethereum_db")
				httpPort = cfg.EthHTTPPort
				grpcPort = cfg.EthGrpcPort
				chainNode, err = node.NewEthNode(ctx, cfg.EthNodeAddress, cfg.NodeChanSize, cfg.RequestPerSecond, log)
				if err != nil {
					log.Error("Error creating eth node", err)
					log.Fatalf(err.Error())
				}
			case "bsc":
				dbPath = fmt.Sprintf("%s/%s", cfg.DatabasePath, "/binance_db")
				httpPort = cfg.BscHTTPPort
				grpcPort = cfg.BscGrpcPort
				chainNode, err = node.NewBscNode(ctx, cfg.BscNodeAddress, cfg.NodeChanSize, cfg.RequestPerSecond, log)
				if err != nil {
					log.Error("Error creating bsc node", err)
					log.Fatalf(err.Error())
				}
			default:
				log.Fatalf(errors.New("chain flag only accept eth and bsc").Error())

			}
			chainDb, err := db.NewLevelDB[model.Account](dbPath)
			if err != nil {
				log.Error("Error opening chainDb", err)
				log.Fatalf(err.Error())
			}
			defer func() {
				err := chainDb.Close()
				if err != nil {
					log.Error("Error closing chainDb", err)
				}
			}()
			startChain(ctx, cfg, chainDb, httpPort, grpcPort, chainNode, log, g)
		}
		if err := g.Wait(); err != nil {
			log.Error("Error in sync ", err)
		}
		if err != nil {
			log.Fatalf("Error running command: %v", err)
		}
	},
}

func startChain(ctx context.Context, config *config.Config, chainDb *db.LvlClient[model.Account], httpPort uint16, grpcPort uint16, chainNode node.Node, log *utils.ZapLogger, g *errgroup.Group) {
	if config.GRPC {
		grpcService := grpc.NewGrpc(chainDb, log, grpcPort)
		g.Go(func() error {
			return grpcService.Run()
		})
	}
	httpService := http.New(chainDb, log, httpPort)
	go func() {
		if err := httpService.Run(); err != nil {
			log.Errorw("Error in service server", "error", err)
		}
	}()
	syncNode, err := sync.New(ctx, chainNode, chainDb, log)
	if err != nil {
		log.Error("Error creating eth sync", err)
		log.Fatalf(err.Error())
	}

	g.Go(func() error {
		return syncNode.Start()
	})
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&cfg.Chain, chainF, "c", defaultChain, chainUsage)
	startCmd.Flags().StringVar(&cfg.EthNodeAddress, ethNodeAddressF, defaultEthNodeAddress, nodeAddressUsage)
	startCmd.Flags().StringVar(&cfg.BscNodeAddress, bscNodeAddressF, defaultBscAddress, nodeAddressUsage)
	startCmd.Flags().Var(&defaultLogLevel, logLevelF, logLevelFlagUsage)
	startCmd.Flags().String(dbPathF, "./dbStore", dbPathUsage)
	startCmd.Flags().Uint16Var(&cfg.EthHTTPPort, ethHTTPPortF, defaultEthHttpPort, ethHTTPPortUsage)
	startCmd.Flags().Uint16Var(&cfg.EthGrpcPort, ethGrpcPortF, defaultEthGrpcPort, ethGrpcPortUsage)
	startCmd.Flags().Uint16Var(&cfg.BscHTTPPort, bscHTTPPortF, defaultBscHttpPort, bscHTTPPortUsage)
	startCmd.Flags().Uint16Var(&cfg.BscGrpcPort, bscGrpcPortF, defaultBscGrpcPort, bscGrpcPortUsage)
	startCmd.Flags().BoolVar(&cfg.GRPC, grpcF, defaultGRPC, grpcUsage)

	startCmd.Flags().String(nodeChanSizeF, defaultNodeChanSize, nodeChanSizeUsage)
	startCmd.Flags().String(requestPerSecondF, defaultRequestPerSecond, requestPerSecondUsage)
}
