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
	router := handlers.NewRouter(server.handler)
	httpServer := &http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}
	log.Fatal(httpServer.ListenAndServe())
}
