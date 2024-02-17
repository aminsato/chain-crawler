package config

import "chain-crawler/utils"

type Config struct {
	HTTP           bool   `mapstructure:"http"`
	EthHTTPPort    uint16 `mapstructure:"eth-http-port"`
	EthGrpcPort    uint16 `mapstructure:"eth-grpc-port"`
	BscHTTPPort    uint16 `mapstructure:"bsc-http-port"`
	BscGrpcPort    uint16 `mapstructure:"bsc-grpc-port"`
	DatabasePath   string `mapstructure:"db-path"`
	EthNodeAddress string `mapstructure:"eth-node-address"`
	BscNodeAddress string `mapstructure:"bsc-node-address"`

	LogLevel         utils.LogLevel `mapstructure:"log-level"`
	Colour           bool           `mapstructure:"colour"`
	NodeChanSize     int            `mapstructure:"node-chan-size"`
	RequestPerSecond int            `mapstructure:"rps"`
	Chain            string         `mapstructure:"chain"`
	GRPC             bool           `mapstructure:"grpc"`
}
