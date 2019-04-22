package config

import "github.com/lillilli/logger"

// Config - service configuration
type Config struct {
	WS WSServer

	FrontendDistPath string
	WatchDir         string

	Log logger.Params
}

// WSServer - ws server configuration
type WSServer struct {
	Host string `default:"0.0.0.0"`
	Port int    `default:"8081"`
}
