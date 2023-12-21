package config

import (
	"flag"

	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
	"github.com/sN00b1/yp-url-shortener/internal/app/server"
)

type Config struct {
	ServerConfig  *server.ServerConfig
	HandlerConfig *handlers.HandlerConfig
}

func New() *Config {
	addrFlag := flag.String("a", "", "host addr")
	urlFlag := flag.String("b", "", "handler base url")
	flag.Parse()
	return &Config{
		ServerConfig:  server.NewServerConfig(*addrFlag),
		HandlerConfig: handlers.NewHandlerConfig(*urlFlag),
	}
}
