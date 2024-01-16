package config

import "chain-crawler/utils"

type Config struct {
	HTTP             bool           `mapstructure:"http"`
	HTTPPort         uint16         `mapstructure:"http-port"`
	DatabasePath     string         `mapstructure:"db-path"`
	NodeAddress      string         `mapstructure:"node-address"`
	LogLevel         utils.LogLevel `mapstructure:"log-level"`
	Colour           bool           `mapstructure:"colour"`
	NodeChanSize     int            `mapstructure:"node-chan-size"`
	RequestPerSecond int            `mapstructure:"rps"`
}
