package main

import (
	"flag"

	"github.com/sN00b1/yp-url-shortener/internal/app/config"
	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
	"github.com/sN00b1/yp-url-shortener/internal/app/server"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
)

func main() {
	addrFlag := flag.String("a", "", "host addr")
	urlFlag := flag.String("b", "", "handler base url")
	flag.Parse()
	addr := config.NewServerConfig(*addrFlag)
	url := config.NewHandlerConfig(*urlFlag)
	g := tools.HashGenerator{}
	s := storage.NewStorage()
	h := handlers.NewHandler(s, g, url)
	server := server.NewServer(h, &addr)
	server.Run()
}
