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
	hostFlag := flag.String("host", "", "host ip address")
	portFlag := flag.String("port", "", "host port")
	flag.Parse()
	c := config.NewConfig(*hostFlag, *portFlag)
	g := tools.HashGenerator{}
	s := storage.NewStorage()
	h := handlers.NewHandler(s, g, c)
	server := server.NewServer(h, &c)
	server.Run()
}
