package server

import (
	"os"
)

type ServerConfig struct {
	ServerAddr string
}

func NewServerConfig(addrFlag string) *ServerConfig {
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
	return &ServerConfig{
		ServerAddr: addr,
	}
}
