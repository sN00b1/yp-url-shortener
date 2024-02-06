package config

import (
	"flag"

	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
	"github.com/sN00b1/yp-url-shortener/internal/app/server"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
)

type Config struct {
	ServerConfig  *server.ServerConfig
	HandlerConfig *handlers.HandlerConfig
	StorageConfig *storage.StorageConfig
}

func New() *Config {
	addrFlag := flag.String("a", "", "host addr")
	urlFlag := flag.String("b", "", "handler base url")
	pathFlag := flag.String("f", "", "file path to save storage info")
	flag.Parse()
	return &Config{
		ServerConfig:  server.NewServerConfig(*addrFlag),
		HandlerConfig: handlers.NewHandlerConfig(*urlFlag),
		StorageConfig: storage.NewStorageConfig(*pathFlag),
	}
}
