package main

import (
	"log"

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
	str := cfg.StorageConfig

	g := tools.HashGenerator{}

	var err error
	var s handlers.Repository
	if str.DBInfo != "" {
		s, err = storage.NewDBStorage(str.DBInfo)
	} else {
		s, err = storage.NewRAMFileStorage(str)
	}

	if err != nil {
		log.Println(err)
	}
	defer s.DeInit()

	h := handlers.NewHandler(s, &g, *url)
	server := server.NewServer(h, addr)
	server.Run()
}
