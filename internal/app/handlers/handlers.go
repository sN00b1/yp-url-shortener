package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
)

type Handler struct {
	storage   storage.Repository
	generator tools.Generator
	mux       *chi.Mux
}

func NewHandler(s storage.Repository, g tools.Generator) *Handler {
	return &Handler{
		storage:   s,
		generator: g,
	}
}

func (handler *Handler) SaveURL(writer http.ResponseWriter, request *http.Request) {
	url, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	hash, err := handler.generator.MakeHash(string(url))
	if hash == "" {
		http.Error(writer, "cannot generate url", http.StatusInternalServerError)
		return
	}
	handler.storage.Save(string(url), hash)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	_, err = writer.Write([]byte("http://localhost:8080/" + hash))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler *Handler) GetURL(writer http.ResponseWriter, request *http.Request) {
	hash := strings.TrimPrefix(request.URL.Path, "/")
	url, err := handler.storage.Get(hash)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if url == "" {
		http.Error(writer, "cant find url by hash", http.StatusNotFound)
	}

	http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
}

func NewRouter(handler *Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Route("/", func(router chi.Router) {
		router.Get("/{id}", handler.GetURL)
		router.Post("/", handler.SaveURL)
	})
	return router
}
