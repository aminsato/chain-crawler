package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

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
	defaultRequestPerSecond = "1"
	defaultNodeAddress      = "https://mainnet.infura.io/v3/your token"

	logLevelF             = "log-level"
	NodeAddressF          = "node-address"
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

		db, err := db.NewLevelDB[model.Account](cfg.DatabasePath)
		if err != nil {
			log.Error("Error opening db", err)
			return err
		}

		defer func() {
			err := db.Close()
			if err != nil {
				log.Error("Error closing db", err)
			}
		}()
		httpService := http.New(db, log, cfg.HTTPPort)
		go func() {
			if err := httpService.Run(); err != nil {
				log.Errorw("Error in http server", "error", err)
			}
		}()

		node, err := node.NewEthNode(ctx, cfg.NodeAddress, cfg.NodeChanSize, cfg.RequestPerSecond, log)
		if err != nil {
			return err
		}
		s, err := sync.New(ctx, node, db, log)
		if err != nil {
			return err
		}
		return s.Start()
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

		// TextUnmarshallerHookFunc allows us to unmarshal values that satisfy the
		// encoding.TextUnmarshaller interface (see the LogLevel type for an example).
		return v.Unmarshal(config, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(), mapstructure.StringToTimeDurationHookFunc())))
	}

	var defaultDBPath string
	defaultDBPath, cwdErr = os.Getwd()
	// Use empty string if we can't get the working directory.
	// We don't want to return an error here since that would make `--help` fail.
	// If the error is non-nil and a db path is not provided by the user, we'll return it in PreRunE.
	if cwdErr == nil {
		defaultDBPath = filepath.Join(defaultDBPath, "ethereum_db")
	}

	defaultLogLevel := utils.INFO

	ethCmd.Flags().String(NodeAddressF, defaultNodeAddress, nodeAddressUsage)
	ethCmd.Flags().Var(&defaultLogLevel, logLevelF, logLevelFlagUsage)
	ethCmd.Flags().String(dbPathF, defaultDBPath, dbPathUsage)
	ethCmd.Flags().String(httpPortF, defaultHTTPPort, httpPortUsage)
	ethCmd.Flags().String(nodeChanSizeF, defaultNodeChanSize, nodeChanSizeUsage)
	ethCmd.Flags().String(requestPerSecondF, defaultRequestPerSecond, requestPerSecondUsage)
	return ethCmd
}
