package main

import (
	"fmt"
	"net/http"

	"github.com/sN00b1/yp-url-shortener/internal/app/server"
)

func main() {
	server := server.NewServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.ShortenerHandler)
	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		fmt.Println("ListenAndServe error: ", err)
	}
}
