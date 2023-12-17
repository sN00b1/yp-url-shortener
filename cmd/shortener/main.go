package main

import (
	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
	"github.com/sN00b1/yp-url-shortener/internal/app/server"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
)

func main() {
	g := tools.HashGenerator{}
	s := storage.NewStorage()
	h := handlers.NewHandler(s, g)
	server := server.NewServer(h)
	server.Run()
}
