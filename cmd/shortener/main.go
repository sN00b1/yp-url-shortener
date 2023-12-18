package main

import (
	"github.com/sN00b1/yp-url-shortener/internal/app/config"
	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
	"github.com/sN00b1/yp-url-shortener/internal/app/server"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
)

func main() {
	c := config.NewConfig()
	g := tools.HashGenerator{}
	s := storage.NewStorage()
	h := handlers.NewHandler(s, g, c)
	server := server.NewServer(h)
	server.Run()
}
