package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
)

type Handler struct {
	storage storage.Storage
}

func NewHandler() *Handler {
	return &Handler{
		storage: *storage.NewStorage(),
	}
}

func (handler *Handler) SaveURL(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusCreated)
	url, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	hash := tools.MakeHash(url)
	if hash == "" {
		http.Error(writer, "Cannot generate url", http.StatusInternalServerError)
		return
	}
	handler.storage.Save(string(url), hash)
	writer.Write([]byte("http://localhost:8080/" + hash))
	fmt.Println("added url: ", hash)
}

func (handler *Handler) GetURL(writer http.ResponseWriter, request *http.Request) {
	hash := request.URL.Path
	url, err := handler.storage.Get(hash)
	fmt.Println("Parsed hash: ", hash)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if url == "" {
		http.Error(writer, "Can't find url by hash", http.StatusNotFound)
	}
	http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
}
