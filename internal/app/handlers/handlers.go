package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
)

type Handler struct {
	storage   storage.Repository
	generator tools.Generator
}

func NewHandler(s storage.Repository, g tools.Generator) *Handler {
	return &Handler{
		storage:   s,
		generator: g,
	}
}

func (handler *Handler) ShortenerHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			handler.getURL(writer, request)
		case http.MethodPost:
			handler.saveURL(writer, request)
		default:
			http.Error(writer, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
	}
}

func (handler *Handler) saveURL(writer http.ResponseWriter, request *http.Request) {
	url, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	hash, err := handler.generator.MakeHash(string(url))
	if hash == "" {
		http.Error(writer, "Cannot generate url", http.StatusInternalServerError)
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

func (handler *Handler) getURL(writer http.ResponseWriter, request *http.Request) {
	hash := strings.TrimPrefix(request.URL.Path, "/")
	url, err := handler.storage.Get(hash)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if url == "" {
		http.Error(writer, "Can't find url by hash", http.StatusNotFound)
	}

	http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
}
