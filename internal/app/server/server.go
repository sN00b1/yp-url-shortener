package server

import (
	"log"
	"net/http"

	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
)

type Server struct {
	handler *handlers.Handler
}

func NewServer(h *handlers.Handler) *Server {
	return &Server{
		handler: h,
	}
}

func (server *Server) Run() {
	httpServer := &http.Server{
		Addr:    "localhost:8080",
		Handler: server.Handler(),
	}
	log.Fatal(httpServer.ListenAndServe())
}

func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	h := server.handler.ShortenerHandler()
	mux.HandleFunc("/", h)
	return mux
}
