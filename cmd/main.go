package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"chain-crawler/config"
	"chain-crawler/db"
	"chain-crawler/http"
	"chain-crawler/model"
	"chain-crawler/sync"
	"chain-crawler/utils"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

import (
	"context"
	"syscall"

	"chain-crawler/node"
)

/*
1. cleanup
2. fetch contract address ? Yes/No => if yes, update node
3. check lint and unit test
4. push to github in chaincrw repo
5. cleanup main => command version
6. docker-compose => restart on failure
*/
const (
	dbPathF                 = "db-path"
	dbPathUsage             = "Location of the database files."
	defaultHTTPPort         = "1080"
	defaultNodeChanSize     = "10"
	defaultRequestPerSecond = "10"
	defaultEthNodeAddress   = "your-api-key"
	defaultBscAddress       = "your-api-key"

	logLevelF             = "log-level"
	ethNodeAddressF       = "eth-node-address"
	bscNodeAddressF       = "bsc-node-address"
	httpPortF             = "http-port"
	nodeChanSizeF         = "node-chan-size"
	requestPerSecondF     = "rps"
	nodeChanSizeUsage     = "The size of the channel that will be used to communicate with the eth node."
	httpPortUsage         = "The httpPortF on which the HTTP server will listen for requests."
	Version               = "0.0.1"
	logLevelFlagUsage     = "Options: debug, info, warn, error."
	nodeAddressUsage      = "The address of the node to connect to."
	requestPerSecondUsage = "Maximum number of requests per second for gateway endpoints"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	go func() {
		<-quit
		cancel()
	}()
	cfg := new(config.Config)
	cmd := NewCmd(cfg, func(cmd *cobra.Command, _ []string) error {
		log, err := utils.NewZapLogger(cfg.LogLevel, cfg.Colour)
		if err != nil {
			return err
		}
		log.Infow("Start crawler")

		//initMigration(cfg.DatabasePath)
		//
		ethDb, err := db.NewLevelDB[model.Account](cfg.DatabasePath + "/ethereum_db")
		if err != nil {
			log.Error("Error opening ethDb", err)
			return err
		}
		bscDb, err := db.NewLevelDB[model.Account](cfg.DatabasePath + "/binance_db")
		if err != nil {
			log.Error("Error opening bscDb", err)
			return err
		}

		defer func() {
			err := ethDb.Close()
			if err != nil {
				log.Error("Error closing ethDb", err)
			}
			err = bscDb.Close()
			if err != nil {
				log.Error("Error closing bscDb", err)
			}
		}()
		httpService := http.New(ethDb, log, cfg.HTTPPort)
		go func() {
			if err := httpService.Run(); err != nil {
				log.Errorw("Error in http server", "error", err)
			}
		}()

		ethNode, err := node.NewEthNode(ctx, cfg.EthNodeAddress, cfg.NodeChanSize, cfg.RequestPerSecond, log)
		if err != nil {
			return err
		}
		bdcNode, err := node.NewBscNode(ctx, cfg.BscNodeAddress, cfg.NodeChanSize, cfg.RequestPerSecond, log)
		if err != nil {
			return err
		}
		syncEth, err := sync.New(ctx, ethNode, ethDb, log)
		if err != nil {
			return err
		}
		syncBsc, err := sync.New(ctx, bdcNode, bscDb, log)
		_ = syncEth
		if err != nil {
			return err
		}

		g.Go(func() error {
			return syncEth.Start()
		})
		g.Go(func() error {
			return syncBsc.Start()
		})
		if err := g.Wait(); err != nil {
			fmt.Println(err)
		}
		return err
	})
	if err := cmd.ExecuteContext(ctx); err != nil {
		log, err := utils.NewZapLogger(cfg.LogLevel, cfg.Colour)
		log.Errorw("Error running command", "error", err)
		os.Exit(1)
	}
}

func NewCmd(config *config.Config, run func(*cobra.Command, []string) error) *cobra.Command {
	ethCmd := &cobra.Command{
		Use:     "eth [flags]",
		Short:   "eth crawler",
		Version: Version,
		RunE:    run,
	}

	var cwdErr error

	ethCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		if cwdErr != nil && config.DatabasePath == "" {
			return fmt.Errorf("find current working directory: %v", cwdErr)
		}

		v := viper.New()

		v.AutomaticEnv()
		v.SetEnvPrefix("Eth")
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		if err := v.BindPFlags(cmd.Flags()); err != nil {
			return nil
		}

		return v.Unmarshal(config, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(), mapstructure.StringToTimeDurationHookFunc())))
	}

	var defaultDBPath string
	defaultDBPath, cwdErr = os.Getwd()
	if cwdErr == nil {
		defaultDBPath = filepath.Join(defaultDBPath, "dbStore")
	}

	defaultLogLevel := utils.INFO

	ethCmd.Flags().String(ethNodeAddressF, defaultEthNodeAddress, nodeAddressUsage)
	ethCmd.Flags().String(bscNodeAddressF, defaultBscAddress, nodeAddressUsage)
	ethCmd.Flags().Var(&defaultLogLevel, logLevelF, logLevelFlagUsage)
	ethCmd.Flags().String(dbPathF, defaultDBPath, dbPathUsage)
	ethCmd.Flags().String(httpPortF, defaultHTTPPort, httpPortUsage)
	ethCmd.Flags().String(nodeChanSizeF, defaultNodeChanSize, nodeChanSizeUsage)
	ethCmd.Flags().String(requestPerSecondF, defaultRequestPerSecond, requestPerSecondUsage)
	return ethCmd
}
