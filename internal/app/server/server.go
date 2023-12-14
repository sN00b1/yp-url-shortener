package server

import (
	"net/http"

	"github.com/sN00b1/yp-url-shortener/internal/app/handlers"
)

type Server struct {
	handler *handlers.Handler
}

func NewServer() *Server {
	return &Server{
		handler: handlers.NewHandler(),
	}
}

func (server *Server) ShortenerHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		server.handler.GetURL(writer, request)
	case http.MethodPost:
		server.handler.SaveURL(writer, request)
	default:
		http.Error(writer, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
}
