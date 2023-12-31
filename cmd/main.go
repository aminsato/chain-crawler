package main

import (
	"ethereum-crawler/config"
	"ethereum-crawler/utils"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
)

/*
1. last height , last transaction
2.
*/
import (
	"context"
	"ethereum-crawler/node"
	"syscall"
)

const (
	dbPathF           = "db-path"
	dbPathUsage       = "Location of the database files."
	logLevelF         = "log-level"
	configF           = "config"
	Version           = "0.0.1"
	logLevelFlagUsage = "Options: debug, info, warn, error."
)

/*
TODO: read config from env variables
*/

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
		fmt.Println("Start crawler")
		node, err := node.New(cfg, ctx)
		// pass context
		//node.StartFetch()
		//node.Subscribe(1, 1000000)
		go node.FetchBlocks(10000000)
		node.FetchTransactions()
		if err != nil {
			return err
		}
		return nil

	})
	if err := cmd.ExecuteContext(ctx); err != nil {
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

	var cfgFile string
	var cwdErr error

	ethCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		if cwdErr != nil && config.DatabasePath == "" {
			return fmt.Errorf("find current working directory: %v", cwdErr)
		}

		v := viper.New()
		if cfgFile != "" {
			v.SetConfigType("yaml")
			v.SetConfigFile(cfgFile)
			if err := v.ReadInConfig(); err != nil {
				return err
			}
		}

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

	ethCmd.Flags().String("node", "http://localhost:8545", "node address")
	ethCmd.Flags().Var(&defaultLogLevel, logLevelF, logLevelFlagUsage)
	ethCmd.Flags().String(dbPathF, defaultDBPath, dbPathUsage)

	return ethCmd
}
