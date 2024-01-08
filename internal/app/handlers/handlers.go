package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Handler struct {
	storage   Repository
	generator Generator
	mux       *chi.Mux
	cfg       HandlerConfig
}

func NewHandler(s Repository, g Generator, c HandlerConfig) *Handler {
	return &Handler{
		storage:   s,
		generator: g,
		mux:       chi.NewMux(),
		cfg:       c,
	}
}

func (handler *Handler) Shorten(writer http.ResponseWriter, request *http.Request) {
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
	result := fmt.Sprintf("%s/%s", handler.cfg.HandlerURL, hash)
	_, err = writer.Write([]byte(result))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler *Handler) Expand(writer http.ResponseWriter, request *http.Request) {
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

func (handler *Handler) ShortenFromJSON(writer http.ResponseWriter, request *http.Request) {
	type inputStruct struct {
		OriginalURL string `json:"url"`
	}
	var input inputStruct
	type outputStruct struct {
		Result string `json:"result"`
	}
	var output outputStruct
	var buf bytes.Buffer

	_, err := buf.ReadFrom(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}
	if err = json.Unmarshal(buf.Bytes(), &input); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("12345678:", input.OriginalURL)
	hash, err := handler.generator.MakeHash(string(input.OriginalURL))
	if hash == "" {
		http.Error(writer, "cannot generate url", http.StatusInternalServerError)
		return
	}
	handler.storage.Save(string(input.OriginalURL), hash)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	output.Result = fmt.Sprintf("%s/%s", handler.cfg.HandlerURL, hash)

	resp, err := json.Marshal(output)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	_, err = writer.Write([]byte(resp))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func NewRouter(handler *Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	//router.Use(loggin.LogginResponse)
	router.Route("/", func(router chi.Router) {
		router.Get("/{id}", handler.Expand)
		router.Post("/", handler.Shorten)
		router.Post("/api/shorten", handler.ShortenFromJSON)
	})
	return router
}
