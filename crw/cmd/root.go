package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"chain-crawler/config"
	"chain-crawler/utils"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfg             *config.Config
	cwdErr          error
	defaultLogLevel = utils.INFO
)

var rootCmd = &cobra.Command{
	Use:     "crw",
	Short:   "crw (Chain Crawler) is A versatile and efficient tool designed for exploring and extracting information from blockchain networks",
	Long:    ``,
	Version: Version,
	PreRun: func(cmd *cobra.Command, args []string) {
		if cwdErr != nil && cfg.DatabasePath == "" {
			log.Fatal(fmt.Errorf("find current working directory: %v", cwdErr))
		}

		v := viper.New()

		v.AutomaticEnv()
		v.SetEnvPrefix("crw")
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		if err := v.BindPFlags(cmd.Flags()); err != nil {
			return
		}
		err := v.Unmarshal(cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(), mapstructure.StringToTimeDurationHookFunc())))
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cfg = new(config.Config)

	rootCmd.Flags().Var(&defaultLogLevel, logLevelF, logLevelFlagUsage)
	rootCmd.Flags().String(dbPathF, "./dbStore", dbPathUsage)

	rootCmd.Flags().String(nodeChanSizeF, defaultNodeChanSize, nodeChanSizeUsage)
	rootCmd.Flags().String(requestPerSecondF, defaultRequestPerSecond, requestPerSecondUsage)
}
