package main

import (
	"github.com/sN00b1/yp-url-shortener/internal/app/config"
	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
	"github.com/sN00b1/yp-url-shortener/internal/app/loggin"
	"github.com/sN00b1/yp-url-shortener/internal/app/server"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
	"go.uber.org/zap"
)

func main() {
	loggin.Initialize("debug")
	cfg := config.New()
	addr := cfg.ServerConfig
	url := cfg.HandlerConfig
	str := cfg.StorageConfig

	g := tools.HashGenerator{}

	var s handlers.Repository
	if str.DBInfo != "" {
		s, _ = storage.NewDBStorage(str.DBInfo)
	} else {
		s, _ = storage.NewRAMFileStorage(str)
	}

	loggin.Log.Debug("Storage config:", zap.String("cfg:", str.DBInfo))

	defer s.DeInit()

	h := handlers.NewHandler(s, &g, *url)
	server := server.NewServer(h, addr)
	server.Run()
}
