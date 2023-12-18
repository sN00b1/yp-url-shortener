package config

import (
	"os"
)

type ServerConfig struct {
	ServerAddr string
}

type HandlerConfig struct {
	HandlerURL string
}

func NewHandlerConfig(urlFlag string) HandlerConfig {
	var url string
	urlOS := os.Getenv("BASE_URL")
	if urlFlag != "" {
		url = urlFlag
	}
	if urlOS != "" {
		url = urlOS
	}
	if url == "" {
		url = "http://localhost:8080"
	}
	return HandlerConfig{
		HandlerURL: url,
	}
}

func NewServerConfig(addrFlag string) ServerConfig {
	var addr string
	addrOS := os.Getenv("SERVER_ADDRESS")
	if addrFlag != "" {
		addr = addrFlag
	}
	if addrOS != "" {
		addr = addrOS
	}
	if addr == "" {
		addr = "localhost:8080"
	}
	return ServerConfig{
		ServerAddr: addr,
	}
}
