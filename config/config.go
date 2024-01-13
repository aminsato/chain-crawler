package config

import "ethereum-crawler/utils"

/*
TODO: Define config
1. http address for node
2. ws address for node
*/
type Config struct {
	HTTP          bool           `mapstructure:"http"`
	HTTPHost      string         `mapstructure:"http-host"`
	HTTPPort      uint16         `mapstructure:"http-port"`
	Websocket     bool           `mapstructure:"ws"`
	WebsocketHost string         `mapstructure:"ws-host"`
	WebsocketPort uint16         `mapstructure:"ws-port"`
	DatabasePath  string         `mapstructure:"db-path"`
	NodeAddress   string         `mapstructure:"node-address"`
	Node          string         `mapstructure:"node1"`
	LogLevel      utils.LogLevel `mapstructure:"log-level"`
	Colour        bool           `mapstructure:"colour"`
}
