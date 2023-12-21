package main

import (
	"github.com/sN00b1/yp-url-shortener/internal/app/config"
	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
	"github.com/sN00b1/yp-url-shortener/internal/app/server"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
)

func main() {
	cfg := config.New()
	addr := cfg.ServerConfig
	url := cfg.HandlerConfig
	g := tools.HashGenerator{}
	s := storage.NewStorage()
	h := handlers.NewHandler(s, &g, *url)
	server := server.NewServer(h, addr)
	server.Run()
}
