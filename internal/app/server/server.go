package server

import (
	"log"
	"net/http"

	"github.com/sN00b1/yp-url-shortener/internal/app/config"
	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
)

type Server struct {
	handler *handlers.Handler
	cfg     *config.Config
}

func NewServer(h *handlers.Handler, cfg *config.Config) *Server {
	return &Server{
		handler: h,
		cfg:     cfg,
	}
}

func (server *Server) Run() {
	router := handlers.NewRouter(server.handler)
	addr := server.cfg.Host + ":" + server.cfg.Port
	httpServer := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	log.Fatal(httpServer.ListenAndServe())
}
